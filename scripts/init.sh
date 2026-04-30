#!/usr/bin/env bash
set -euo pipefail

DB_HOST="${DATABASE_HOST:-db}"
DB_PORT="${DATABASE_PORT:-5432}"
DB_USER="${DATABASE_USER:-postgres}"
DB_PASSWORD="${DATABASE_PASSWORD:-postgres}"
DB_NAME="${DATABASE_NAME:-gobid}"

echo "Waiting for database to be ready..."
until PGPASSWORD="$DB_PASSWORD" pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres; do
  sleep 1
done

echo "Creating database if it does not exist..."
if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1; then
  PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "CREATE DATABASE \"$DB_NAME\";"
fi

echo "Applying migrations..."
cd /app
tern migrate --migrations ./migrations --config ./tern.conf

echo "Starting application..."
exec ./bin/go-order-book
