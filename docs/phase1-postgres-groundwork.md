# Phase 1 Postgres Groundwork

- Move MVP data store from in-memory to Postgres backing for core entities: Tenants, Users, Sessions, RefreshTokens, Policies, Translations.
- Rationale: Enables real persistence, constraints, and migration readiness toward per-tenant schemas if required.
- Groundwork artifacts:
  - 001_init_schema.sql (core tables) – located in migrations/
  - 002_seed_data.sql – baseline seed data – located in migrations/
- Plan for Phase 1 integration:
  - Introduce a lightweight Postgres store layer (internal/db) and bridge MVP services to PostgreSQL.
  - Implement minimal CRUD operations for Tenants and Users; expose through existing MVP REST endpoints.
  - Establish a migration runner to apply 001 and 002 in test/staging/production as part of CI/CD.
