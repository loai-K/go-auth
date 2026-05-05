#!/usr/bin/env bash
set -euo pipefail

HOST=${AUTH_API_HOST:-http://localhost:8080}

echo "Running MVP end-to-end contract tests against ${HOST}"

curl -sSf "$HOST/healthz" >/dev/null
curl -sSf -X POST "$HOST/auth/authorize" >/dev/null
curl -sSf "$HOST/token?tenant_id=tenant1" >/dev/null
curl -sSf -X POST "$HOST/token/introspect" -H "Content-Type: application/json" -d '{"token":"dummy"}' >/dev/null
curl -sSf "$HOST/tenants/info?tenant_id=tenant1" >/dev/null
curl -sSf -X POST "$HOST/users" -H "Content-Type: application/json" -d '{"email":"eve@example.com","tenant_id":"tenant1"}' >/dev/null

echo "All contract tests passed"
