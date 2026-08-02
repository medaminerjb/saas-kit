# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-08-02

### Added
- **OIDC Provider & Federation**
  - OIDC Provider integration via `zitadel/oidc/v3` with complete custom `op.Storage` implementation.
  - Social login support for Google and GitHub (OAuth2 relying party with CSRF protection).
  - Production-ready dark-themed Login and Consent UI templates.
  - OIDC client CRUD database queries for `oidc_clients`.
  - Identity provider and external account database queries for federated logins.
- **Identity & Security Hardening**
  - Account lockout mechanism that automatically locks user accounts (`UserStatusLocked`) after 5 unsuccessful login attempts.
  - Thread-safe IP-based token bucket rate limiter middleware for authentication endpoints.
  - Security headers middleware enforcing HSTS, CSP, X-Frame-Options, and other secure defaults.
  - User domain model with 6-state lifecycle (active, disabled, locked, pending, invited, deleted).
  - Argon2id password hashing (PHC format, configurable parameters).
  - Envelope encryption (AES-256-GCM) with HKDF key derivation for database secrets.
  - HMAC token hashing for secure token validation.
- **Core Platform & Infrastructure**
  - Go project scaffold utilizing chi router, pgx connection pool, sqlc, koanf, and slog.
  - PostgreSQL schema database migrations: users, sessions, OIDC, tokens, audit, MFA (reserved).
  - Event publisher interface with Log, Multi, and Noop implementations.
  - Audit log database persistence via `AuditPublisher` with context-injected IP/User-Agent.
  - Client info middleware for automatic IP and User-Agent context propagation.
  - Background job scheduler (ticker-based, graceful shutdown) for session and token cleanup.
  - Multi-stage production Dockerfile and Docker Compose configuration.
  - CI pipeline via GitHub Actions with linting, unit/integration testing, and Docker builds.
  - Security scanning using `govulncheck` integrated into the CI workflow.
  - OSS governance documentation (CONTRIBUTING, CODE_OF_CONDUCT, SECURITY, GOVERNANCE, SUPPORT, PULL_REQUEST_TEMPLATE).
  - Architecture Decision Records (ADRs) for sqlc, asymmetric JWT, koanf, zitadel/oidc.
  - Apache 2.0 license.
- **Testing & Verification**
  - Comprehensive unit test coverage for `UserService`, `AuthService`, token validation, and security middleware.

### Fixed
- Aligned `sqlc` queries for `oidc_clients` with database schema (replaced non-existent columns with actual schema columns).
- Aligned `sqlc` queries for `identity_providers` with database schema.
- Improved `audit_logs` sqlc query parameter naming to use named `ip_address` parameter.

[unreleased]: https://github.com/medaminerjb/saas-kit/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/medaminerjb/saas-kit/releases/tag/v0.1.0
