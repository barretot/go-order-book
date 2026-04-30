-- Write your migrate up statements here

ALTER TABLE wallet_assets
  ADD CONSTRAINT wallet_assets_user_id_fkey
  FOREIGN KEY (user_id)
  REFERENCES users(id)
  ON DELETE CASCADE;

ALTER TABLE orders
  ADD CONSTRAINT orders_user_id_fkey
  FOREIGN KEY (user_id)
  REFERENCES users(id)
  ON DELETE RESTRICT;

---- create above / drop below ----

ALTER TABLE orders
  DROP CONSTRAINT IF EXISTS orders_user_id_fkey;

ALTER TABLE wallet_assets
  DROP CONSTRAINT IF EXISTS wallet_assets_user_id_fkey;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
