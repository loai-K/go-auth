# GoAuth — Production‑Grade Auth Platform

GoAuth is a production‑grade authentication and authorization platform built in Go. This repository contains a Phase 1 MVP scaffold (in‑memory MVP server, REST/gRPC surface, OpenAPI/Proto contracts, and initial migration groundwork) and a plan for progression to a Postgres‑backed, multi‑tenant, policy‑driven platform.

## Quick start
- Prerequisites: Go 1.26+, optionally Docker for containerized runs, and (for Phase 2+) a Postgres instance.
- Build: `make build`
- Run MVP server (in‑memory): `make run` or `go run ./cmd/auth-server`
- Run unit tests: `make test` or `go test ./...`
- Run Docker container (Phase 1 MVP): `docker build -t goauth:phase1-mvp .` followed by `docker run -p 8080:8080 goauth:phase1-mvp`
  - Endpoints (Phase 1 MVP):
    - GET /health
    - POST /auth/authorize
    - POST /token
    - POST /token/revoke
    - POST /token/introspect
    - GET /tenants/info?tenant_id=...
    - POST /users
  - The OpenAPI contract for REST is located at `api/openapi/auth.yaml` and the gRPC surface is at `proto/auth.proto`.

## Environment variables
- PORT: The port number to listen on. Default: 8080
- POSTGRES_DSN: The DSN for the Postgres database. Default: "host=localhost port=5432 user=postgres password=postgres dbname=goauth sslmode=disable" if not set
- JWT_SIGNING_KEY: The secret key used for signing JWT tokens. Default: "your_jwt_signing_key"
USE: "openssl rand -hex 32" or "openssl rand -base64 32"

## Project structure
- cmd/                – main executables (auth server, playgrounds)
- internal/           – core services and primitives
  - store/            – in‑memory domain models and storage scaffolds
  - token/            – token creation surface (MVP)
  - db/               – DB abstractions and skeletons (Postgres groundwork)
  - httpx/            – lightweight HTTP response envelopes
- migrations/         – SQL migration scripts (001_init_schema.sql, 002_seed_data.sql)
- api/ openapi/       – REST contract for Phase 1
- proto/              – Protobuf contracts for Phase 1
- docs/               – design plans and architectural notes
- tests/              – simple end‑to‑end tests and helpers
- Dockerfile          – container image for MVP
- Makefile            – convenience targets (build/test/run/docker)
- .github/ workflows/ ci.yml – CI workflow for builds and tests

## MVP Phase 1 scope
- Implement core authentication flows: OAuth2/OIDC authorization code flow with PKCE, client credentials, and device code (Phase 1 skeleton).
- Basic user/tenant data model with in‑memory storage for rapid iteration.
- Token minting surface, rotating refresh tokens, and secure session placeholders.
- REST/gRPC surfaces for developers and admin console scaffolding.
- Phase 1 migration groundwork toward Postgres (DDL in migrations/001_init_schema.sql).

## Phase 2+ roadmap (high level)
- Move from in‑memory MVP to Postgres‑backed stores; implement real CRUD for tenants/users
- Introduce RBAC/ABAC policy engine (Cerbos recommended) and policy lifecycle
- Add multi‑tenancy isolation options (Row-Level Security or per‑tenant schemas)
- Expand extensibility: webhooks/actions, i18n, admin tooling, SDKs, and auditing
- Improve security posture, observability, and compliance readiness

## How to contribute
- Review issues and PRs, propose small, well scoped changes, and write tests for new features.
- Follow the existing patch format and provide a clear rationale in PR descriptions.

## License
- Project is open for collaboration. See repository for licensing details.

If you’d like more details on any part of the architecture or a different MVP emphasis (e.g., sooner DB migration or a different policy engine), tell me and I’ll adjust the plan.
