-- name: CreateUser :one
INSERT INTO users (name, email )
VALUES ($1, $2)
RETURNING id;

-- name: GetUserById :one
SELECT
  id,
  name,
  email,
  created_at,
  updated_at
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT
  id,
  name,
  email,
  created_at,
  updated_at
FROM users
ORDER BY created_at DESC;
