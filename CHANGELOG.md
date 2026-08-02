# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- OIDC Provider via `zitadel/oidc/v3` with full `op.Storage` implementation.
- Social login support for Google and GitHub (OAuth2 relying party with CSRF protection).
- Login and consent UI (Go templates, dark-themed, production-replaceable).
- `CreateSessionForUser` for passwordless social login sessions.
- OIDC client CRUD queries (sqlc) for `oidc_clients` table.
- Identity provider and external account queries (sqlc) for `identity_providers` and `external_accounts` tables.
- Audit log database persistence via `AuditPublisher` with context-injected IP/User-Agent.
- Client info middleware for automatic IP and User-Agent context propagation.
- OSS governance documentation (CONTRIBUTING, CODE_OF_CONDUCT, SECURITY, GOVERNANCE, SUPPORT).
- GitHub issue and pull request templates.
- Architecture specification (`docs/architecture.md`).

### Fixed
- Aligned `sqlc` queries for `oidc_clients` with database schema (replaced non-existent columns `scope_restrictions`, `is_active`, `id_token_lifetime`, `access_token_lifetime`, `refresh_token_lifetime`, `access_token_type` with actual schema columns).
- Aligned `sqlc` queries for `identity_providers` with database schema (replaced `provider_type` → `protocol`, `discovery_url` → `issuer_url`, `is_active` → `enabled`, `profile_data` → `raw_data`).
- Improved `audit_logs` sqlc query parameter naming (replaced `Column5` with named `ip_address` parameter).

## [0.1.0] — Unreleased (Developer Preview)

### Added
- **Core Platform**
  - Go project scaffold with chi, pgx, sqlc, koanf, and slog.
  - PostgreSQL schema (10 migrations): users, sessions, OIDC, tokens, audit, MFA (reserved).
  - Configuration system via koanf with env vars (`SAASKIT_` prefix) and YAML fallback.
  - PostgreSQL connection pool with health check and configurable pool size.
  - Envelope encryption (AES-256-GCM) with HKDF key derivation and context separation.
  - Event publisher interface with Log, Multi, and Noop implementations.
  - Background job scheduler (ticker-based, graceful shutdown).
  - Docker Compose and multi-stage production Dockerfile.
  - CI pipeline (GitHub Actions): lint, test, build.
  - Architecture Decision Records (ADRs) for sqlc, asymmetric JWT, koanf, zitadel/oidc.
  - Apache 2.0 license.

- **Identity System**
  - User domain model with 6-state lifecycle (active, disabled, locked, pending, invited, deleted).
  - Argon2id password hashing (PHC format, configurable parameters).
  - JWT key management supporting RS256, ES256, and EdDSA.
  - HMAC token hashing (server-secret-bound, not raw SHA-256).
  - Auth service: register, login, refresh, logout with refresh token rotation.
  - Password reset and email verification flows via generic identity token system.
  - User profile CRUD with soft delete support.
  - Session management: list, revoke, and cleanup job.
  - HTTP API with JWT auth middleware, CORS, and request logging.
  - OIDC Discovery endpoint at `/.well-known/openid-configuration`.
  - IdentityManager orchestrator (extensible for MFA/WebAuthn).
  - Audit event publisher logging all identity operations.
  - Unit tests for crypto and token services (16 passing, race detection enabled).

[unreleased]: https://github.com/saaskit/saaskit/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/saaskit/saaskit/releases/tag/v0.1.0
