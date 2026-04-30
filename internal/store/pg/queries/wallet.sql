-- name: CreateWalletAsset :one
INSERT INTO wallet_assets (user_id, instrument, quantity)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, instrument)
DO UPDATE SET
  quantity = wallet_assets.quantity + EXCLUDED.quantity,
  updated_at = now()
RETURNING id;

-- name: DebitWalletAsset :execrows
UPDATE wallet_assets
SET
  quantity = quantity - $3,
  updated_at = now()
WHERE user_id = $1
  AND instrument = $2
  AND quantity >= $3;

-- name: GetWalletByUserId :many
SELECT
  id,
  user_id,
  instrument,
  quantity,
  created_at,
  updated_at
FROM wallet_assets
WHERE user_id = $1
ORDER BY instrument;

-- name: ListWallets :many
SELECT
  id,
  user_id,
  instrument,
  quantity,
  created_at,
  updated_at
FROM wallet_assets
ORDER BY created_at DESC;
