INSERT INTO users (id, name, email)
VALUES
  ('4e4de77c-72f4-46de-bd6f-d743ad24acfa', 'Conta A Seed', 'conta-a.seed@example.com'),
  ('d637dfc8-132d-451e-8485-610118670c53', 'Conta B Seed', 'conta-b.seed@example.com'),
  ('8a1b5c90-7d2e-4f31-9c84-2e60f7b4a901', 'Conta C Seed', 'conta-c.seed@example.com'),
  ('6f2c9a13-4b85-4d9e-9a77-1c4f0a6b8e52', 'Conta D Seed', 'conta-d.seed@example.com')
ON CONFLICT (id)
DO UPDATE SET
  name = EXCLUDED.name,
  email = EXCLUDED.email,
  updated_at = now();

INSERT INTO wallet_assets (user_id, instrument, quantity)
VALUES
  ('4e4de77c-72f4-46de-bd6f-d743ad24acfa', 'BRL', 500000),
  ('d637dfc8-132d-451e-8485-610118670c53', 'BTC', 1),
  ('8a1b5c90-7d2e-4f31-9c84-2e60f7b4a901', 'BRL', 1000000),
  ('6f2c9a13-4b85-4d9e-9a77-1c4f0a6b8e52', 'BTC', 1)
ON CONFLICT (user_id, instrument)
DO UPDATE SET
  quantity = EXCLUDED.quantity,
  updated_at = now();
