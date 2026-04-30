-- Write your migrate up statements here

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  email text UNIQUE NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS wallet_assets (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  instrument TEXT NULL,
  quantity NUMERIC(20, 8) NULL DEFAULT 0,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT wallet_assets_user_instrument_unique UNIQUE (user_id, instrument),
  CONSTRAINT wallet_assets_quantity_non_negative CHECK (quantity >= 0)
);

CREATE TABLE IF NOT EXISTS orders (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  instrument TEXT NOT NULL,
  quantity NUMERIC(20, 8) NULL DEFAULT 0,
  remaining_quantity NUMERIC(20, 8) NOT NULL,
  price NUMERIC(20, 8) NOT NULL,
  side TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'open',

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT orders_quantity_positive CHECK (quantity > 0),
  CONSTRAINT orders_remaining_quantity_non_negative CHECK (remaining_quantity >= 0),
  CONSTRAINT orders_price_positive CHECK (price > 0),
  CONSTRAINT orders_side_valid CHECK (side IN ('buy', 'sell')),
  CONSTRAINT orders_status_valid CHECK (status IN ('open', 'partially_filled', 'filled', 'cancelled'))
);

CREATE TABLE IF NOT EXISTS trades (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  buy_order_id UUID NOT NULL,
  sell_order_id UUID NOT NULL,
  instrument TEXT NOT NULL,
  quantity NUMERIC(20, 8) NOT NULL,
  price NUMERIC(20, 8) NOT NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT trades_buy_order_id_fkey
    FOREIGN KEY (buy_order_id)
    REFERENCES orders(id)
    ON DELETE RESTRICT,
  CONSTRAINT trades_sell_order_id_fkey
    FOREIGN KEY (sell_order_id)
    REFERENCES orders(id)
    ON DELETE RESTRICT,
  CONSTRAINT trades_quantity_positive CHECK (quantity > 0),
  CONSTRAINT trades_price_positive CHECK (price > 0)
);

CREATE INDEX IF NOT EXISTS idx_wallet_assets_user_id ON wallet_assets(user_id);
CREATE INDEX IF NOT EXISTS idx_wallet_assets_instrument ON wallet_assets(instrument);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_instrument ON orders(instrument);
CREATE INDEX IF NOT EXISTS idx_orders_matching ON orders(instrument, side, status, price, created_at);
CREATE INDEX IF NOT EXISTS idx_trades_buy_order_id ON trades(buy_order_id);
CREATE INDEX IF NOT EXISTS idx_trades_sell_order_id ON trades(sell_order_id);

---- create above / drop below ----

DROP INDEX IF EXISTS idx_trades_sell_order_id;
DROP INDEX IF EXISTS idx_trades_buy_order_id;
DROP INDEX IF EXISTS idx_orders_matching;
DROP INDEX IF EXISTS idx_orders_instrument;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_wallet_assets_instrument;
DROP INDEX IF EXISTS idx_wallet_assets_user_id;

DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS wallet_assets;
DROP TABLE IF EXISTS users;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
