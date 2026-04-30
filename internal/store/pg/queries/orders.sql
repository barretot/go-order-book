-- name: CreateOrder :one
INSERT INTO orders (user_id, instrument, quantity, remaining_quantity, price, side)
VALUES ($1, $2, $3, $3, $4, $5)
RETURNING id, user_id, instrument, quantity, remaining_quantity, price, side, status;

-- name: FindSellMatchForBuy :one
SELECT
  id,
  user_id,
  instrument,
  quantity,
  remaining_quantity,
  price,
  side,
  status,
  created_at,
  updated_at
FROM orders
WHERE instrument = $1
  AND side = 'sell'
  AND status IN ('open', 'partially_filled')
  AND remaining_quantity > 0
  AND price <= $2
  AND user_id <> $3
ORDER BY price ASC, created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;

-- name: FindBuyMatchForSell :one
SELECT
  id,
  user_id,
  instrument,
  quantity,
  remaining_quantity,
  price,
  side,
  status,
  created_at,
  updated_at
FROM orders
WHERE instrument = $1
  AND side = 'buy'
  AND status IN ('open', 'partially_filled')
  AND remaining_quantity > 0
  AND price >= $2
  AND user_id <> $3
ORDER BY price DESC, created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;

-- name: DecrementOrderRemaining :one
UPDATE orders
SET
  remaining_quantity = remaining_quantity - $2,
  status = CASE
    WHEN remaining_quantity - $2 = 0 THEN 'filled'
    WHEN remaining_quantity - $2 < quantity THEN 'partially_filled'
    ELSE 'open'
  END,
  updated_at = now()
WHERE id = $1
  AND remaining_quantity >= $2
RETURNING id, user_id, instrument, quantity, remaining_quantity, price, side, status;

-- name: CreateTrade :one
INSERT INTO trades (buy_order_id, sell_order_id, instrument, quantity, price)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: CancelOrder :one
UPDATE orders
SET
  status = 'cancelled',
  updated_at = now()
WHERE id = $1
  AND user_id = $2
  AND status IN ('open', 'partially_filled')
RETURNING id, user_id, instrument, quantity, remaining_quantity, price, side, status;

-- name: GetOrderByUserId :many
SELECT
  id,
  user_id,
  instrument,
  quantity,
  remaining_quantity,
  price,
  side,
  status,
  created_at,
  updated_at
FROM orders
WHERE user_id = $1
ORDER BY instrument;

-- name: GetOrderBookByInstrument :many
SELECT
  id,
  user_id,
  instrument,
  quantity,
  remaining_quantity,
  price,
  side,
  status,
  created_at,
  updated_at
FROM orders
WHERE instrument = $1
ORDER BY
  side,
  CASE WHEN side = 'buy' THEN price END DESC,
  CASE WHEN side = 'sell' THEN price END ASC,
  created_at ASC;

-- name: ListOrders :many
SELECT
  id,
  user_id,
  instrument,
  quantity,
  remaining_quantity,
  price,
  side,
  status,
  created_at,
  updated_at
FROM orders
ORDER BY created_at DESC;
