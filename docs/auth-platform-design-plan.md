# Auth Platform Design & Execution Plan

Executive summary
- Build a production‑grade, multi‑tenant authentication and authorization platform in Go, with OAuth2/OIDC, SAML, RBAC/ABAC, i18n, extensibility hooks, and robust observability. The plan covers MVP scope (Phase 1) and phased enhancements (Phases 2–4), data model sketches, API contracts, and a practical 12–16 week delivery roadmap.

## 1. Goals & guiding principles
- Deliver a secure, scalable, and extensible Auth0‑like platform for SaaS, mobile, and enterprise integrations.
- Prioritize: security, reliability, developer experience, observability, and a clear migration path for tenants.
- Provide REST and gRPC developer surfaces, an admin console scaffold, and non‑disruptive migrations.
- Start with a pragmatic tenancy model and policy engine that can evolve over time.

## 2. Tenancy model (starter) and evolution
- Starter approach (Phase 1): Shared Postgres schema with Row-Level Security (RLS) on tenant_id to ensure data isolation at the row level while keeping operations simple.
- Migration path (Phase 2+): Support per‑tenant schemas/databases if tenant count or data residency requirements demand stronger isolation.
- Tradeoffs:
  - Phase 1: Lower ops burden, faster MVP delivery.
  - Phase 2: Greater isolation, easier backups/compliance, higher orchestration cost.

## 3. Policy engine
- Recommended: Cerbos for RBAC/ABAC policy evaluation, versioned policies, staging, and dry‑run capability.
- Integrate via a dedicated Policy Service that exposes a uniform evaluation API for app services.
- Policy lifecycle: versioning, testing, staging, promotion to production, rollbacks.

## 4. Token, sessions, and MFA strategy
- Access tokens: short‑lived JWTs with audience/subject scoped to tenant.
- Refresh tokens: rotate on use, stored securely (e.g., Redis with token bindings). Ability to revoke tokens.
- MFA: baseline support for TOTP and WebAuthn; plan for backup codes and SMS/EMAIL OTP as needed.
- Device and session management: track device_id, IP, user_agent; support selective logout and multi‑device policies.
- Offline access: optional long‑lived refresh tokens with explicit consent.

## 5. Data stores & phase 1 schema (MVP)
- Relational data (Postgres): Users, Tenants, Sessions, RefreshTokens, Policies, Translations.
- Ephemeral/fast data: Redis for session state, token revocation lists, rate limiting, MFA challenges.
- Events: Kafka or NATS for audit events and policy updates.

Phase 1 schema sketch (core tables)
- Users: id (UUID), tenant_id (FK), user_type, email, email_verified, password_hash, password_last_changed, preferred_language, profile JSONB, status, created_at, updated_at
- Tenants: id (UUID), name, slug, default_language, supported_languages text[], settings JSONB, branding JSONB, created_at, updated_at
- Sessions: id (UUID), user_id (FK), device_id, ip_address, user_agent, login_time, expires_at, last_seen, refresh_token_id
- RefreshTokens: id (UUID), user_id (FK), tenant_id, token_hash, expires_at, revoked_at, created_at
- Policies: id (UUID), tenant_id, version, policy_type (RBAC/ABAC), policy_bundle JSONB, status, created_at, updated_at
- Translations: id (UUID), tenant_id, locale, namespace, key, value, updated_at

## 6. MVP scope (Phase 1)
### Objectives
- Implement OAuth2/OIDC flows: authorization code with PKCE, client credentials, device code.
- Build core identity store with password hashing and baseline MFA scaffolding.
- Token minting, rotating refresh tokens, secure sessions, and basic audit logging.
- REST and gRPC API skeletons; admin UI scaffolding.

### Phase 1 API surface
- REST
  - POST /auth/authorize
  - POST /token
  - POST /token/revoke
  - POST /token/introspect
  - GET /tenants/{tenant_id}/info
  - POST /users (admin) / GET /users/{id}
- gRPC
  - Auth service: Token minting, user lookup
  - User service: Create/get/update user
  - Tenant service: Create/get tenant

## 7. Security baseline
- Password storage: Argon2id (or modern bcrypt with strong cost if Argon2id not available)
- MFA: TOTP and WebAuthn scaffolding; enrollment flow in place
- JWT handling: signed access tokens, short lifetimes; rotating refresh tokens; revocation support
- Auditing: login attempts, token issuance, policy decisions

## 8. Observability & reliability
- Metrics: authentication success/failure, token issuance rates, latency percentiles
- Tracing: OpenTelemetry across services
- Logging: structured logs with trace_id and tenant_id
- Availability: define RPO/RTO and plan for HA deployments

## 9. Architecture overview (textual)
- Core services: Auth, User, Tenant, Policy, I18n, Extensibility, Audit
- Data stores: Postgres (durable), Redis (state/store), Kafka/NATS (events)
- API gateway for routing, rate limiting, and authz checks
- Inter-service communication: gRPC; external API surface: REST with gRPC‑gateway compatibility

## 10. Development, CI/CD, and Ops plan
- Repository layout aligned to services (auth, user, tenant, policy, i18n, audit, extensibility)
- Go modules per service; shared libraries for common auth routines and policy evaluation
- CI: unit/integration tests, API contract validation, security scans
- CD: Kubernetes manifests, Helm or Kustomize, GitOps workflow, non-breaking migrations
- Migrations: non‑blocking, incremental migrations; feature flags for rollout

## 11. Phase 2–3 roadmap (high level)
- Phase 2: RBAC/ABAC enforcement, policy engine integration, per‑tenant schema migration plan if needed
- Phase 3: Extensibility (webhooks, actions), i18n localization pipeline, governance dashboards
- Phase 4: Security hardening, compliance, advanced analytics, DR testing

### Phase 2: Policy engine integration (Cerbos)
- Introduce Cerbos as the policy engine, wire a PolicyService to query for access decisions
- Implement policy versioning, staging, dry-run, and production promotion
- Extend RBAC/ABAC with role bindings by tenant and per-resource attributes
- Phase 2: RBAC/ABAC enforcement, policy engine integration, per‑tenant schema migration plan if needed
- Phase 3: Extensibility (webhooks, actions), i18n localization pipeline, governance dashboards
- Phase 4: Security hardening, compliance, advanced analytics, DR testing

## 12. Risks & mitigations
- MVP scope risk: maintain strict Phase 1 scope; defer non‑critical features
- Data isolation risk: start with RLS; plan per‑tenant schemas if needed
- Policy engine risk: start with Cerbos; implement a test policy suite early
- Compliance risk: address GDPR/SOC2 alignment early in design and maintain policy discipline

## 13. Decision log (highlights)
- Tenancy: start with shared schema + RLS; migrate to per‑tenant schemas later
- Policy engine: Cerbos chosen for MVP; API‑driven policy evaluation
- Token strategy: rotate refresh tokens; short‑lived access tokens; revocation lists
- Protocols: OAuth2/OIDC, SAML; gRPC for internal, REST for external developers

## 14. Milestones & delivery cadence (12–16 weeks)
- Week 1–2: Project setup, repository scaffolding, go.mod, base API contracts
- Week 3–5: Phase 1 MVP: auth flows, User/Tenant models, basic services, REST/gRPC skeletons
- Week 6–8: Token lifecycle, MFA scaffolding, Redis state, basic audit/logging
- Week 9–11: Policy integration (Cerbos), RBAC/ABAC scaffolding, initial i18n hooks
- Week 12–14: Admin console scaffolding, migrations plans, CI/CD pipelines
- Week 15–16: Observability rollout, security hardening, readiness review, DR planning

## 15. Appendix: Glossary
- MVP: Minimal Viable Product
- RLS: Row-Level Security
- MFA: Multi‑Factor Authentication
- OIDC: OpenID Connect
- JWT: JSON Web Token
- AES/Argon2id: cryptographic primitives for password storage

---
This document serves as the consolidated plan and will be iterated as we gather requirements, blockers, and implementation insights. Please confirm any deadlines, compliance requirements, or architectural preferences to tailor the roadmap.
