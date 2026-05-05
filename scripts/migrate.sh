#!/usr/bin/env bash
set -euo pipefail

echo "Migration runner: phase-1 (001_init_schema.sql, 002_seed_data.sql)"

if ! command -v psql >/dev/null 2>&1; then
  echo "psql not found in PATH; skipping actual migrations. This script is for CI/debugging environments with Postgres available."
  exit 0
fi

DSN=${DB_DSN:-"postgres://postgres:postgres@localhost:5432/auth?sslmode=disable"}
echo "Using DSN: ${DSN}"

psql "$DSN" -f migrations/001_init_schema.sql
psql "$DSN" -f migrations/002_seed_data.sql

echo "Migrations applied."
