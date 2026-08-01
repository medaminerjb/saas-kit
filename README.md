# SaaSKit

**The open-source, tenancy-native identity and access foundation for SaaS apps.**

SaaSKit provides the foundational infrastructure every multi-tenant SaaS application needs: authentication, organizations, and access control — done once, done well, and reusable across projects.

> A credible open-source alternative to Keycloak with better developer experience.

## Features (v1.0 — in progress)

- **Email/password authentication** with Argon2id password hashing
- **JWT access tokens** with asymmetric signing (RS256, ES256, EdDSA)
- **Refresh token rotation** with HMAC-hashed storage (database leaks don't compromise tokens)
- **OIDC Provider** — SaaSKit acts as a full OpenID Connect Provider
- **Social login** — Google, GitHub (extensible to any OAuth2/OIDC provider)
- **Multi-tenant ready** — `tenant_id` columns reserved across all tables
- **Audit logging** — append-only identity event audit trail
- **Event-driven architecture** — publisher interface for future Kafka/NATS/Redis integration
- **Background jobs** — automatic cleanup of expired sessions and tokens
- **Envelope encryption** — AES-256-GCM for stored secrets (OAuth client secrets, etc.)
- **Generic token system** — password resets, email verification, invites, magic links — one table

## Quick Start

```bash
# Clone the repository
git clone https://github.com/saaskit/saaskit.git
cd saaskit

# Copy the example env file
cp .env.example .env

# Start PostgreSQL and SaaSKit
docker compose up

# Or run locally (requires PostgreSQL)
make migrate-up
make run-direct
```

## API Endpoints

### Authentication (`/api/v1/`)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/auth/register` | Register with email/password |
| `POST` | `/api/v1/auth/login` | Login |
| `POST` | `/api/v1/auth/refresh` | Refresh access token |
| `POST` | `/api/v1/auth/logout` | Logout (requires auth) |
| `POST` | `/api/v1/auth/forgot-password` | Request password reset |
| `POST` | `/api/v1/auth/reset-password` | Execute password reset |
| `POST` | `/api/v1/auth/verify-email` | Verify email address |

### User Profile (`/api/v1/`, requires auth)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/users/me` | Get current user |
| `PATCH` | `/api/v1/users/me` | Update profile |
| `GET` | `/api/v1/users/me/sessions` | List sessions |
| `DELETE` | `/api/v1/users/me/sessions/{id}` | Revoke session |

### OIDC Provider (`/oidc/`)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/.well-known/openid-configuration` | OIDC Discovery |
| `GET` | `/oidc/authorize` | Authorization endpoint |
| `POST` | `/oidc/token` | Token endpoint |
| `GET` | `/oidc/userinfo` | UserInfo endpoint |
| `GET` | `/oidc/keys` | JWKS endpoint |
| `POST` | `/oidc/revoke` | Token revocation |
| `POST` | `/oidc/introspect` | Token introspection |
| `GET` | `/oidc/login` | Login UI (auth code flow) |

### Social Login (`/oauth2/`)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/oauth2/google/login` | Initiate Google OAuth2 |
| `GET` | `/oauth2/google/callback` | Google OAuth2 callback |
| `GET` | `/oauth2/github/login` | Initiate GitHub OAuth2 |
| `GET` | `/oauth2/github/callback` | GitHub OAuth2 callback |

### System

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness check |
| `GET` | `/ready` | Readiness check (DB connectivity) |

## Architecture

```
cmd/saaskit/           → Entry point, dependency wiring
internal/
  config/              → Configuration (koanf, env + YAML)
  platform/
    database/          → PostgreSQL connection pool (pgx)
    crypto/            → Envelope encryption (AES-256-GCM)
    events/            → Event publisher interface
    jobs/              → Background job scheduler
  identity/
    domain/            → User, Session, Token entities
    crypto/            → Argon2id, JWT keys, HMAC token hashing
    service/           → Auth, User, Token services + IdentityManager
    repository/        → PostgreSQL repositories
    handler/           → HTTP handlers, middleware, router
  oidc/
    provider/          → OpenID Connect Provider (zitadel/oidc)
    relyingparty/      → Social login (Google, GitHub)
  audit/               → Audit logging
  tenant/              → [Reserved] Multi-tenancy
  authorization/       → [Reserved] RBAC + policies
  organizations/       → [Reserved] Organization management
```

## Configuration

Configuration is loaded from environment variables (prefix `SAASKIT_`) with optional YAML fallback. See [`.env.example`](.env.example) for all options.

### Key Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SAASKIT_ENV` | Environment (`development`, `production`) | `development` |
| `SAASKIT_PORT` | HTTP server port | `8080` |
| `SAASKIT_BASE_URL` | Public base URL for callbacks | `http://localhost:8080` |
| `SAASKIT_DATABASE_HOST` | PostgreSQL host | `localhost` |
| `SAASKIT_DATABASE_PORT` | PostgreSQL port | `5432` |
| `SAASKIT_DATABASE_USER` | PostgreSQL user | `saaskit` |
| `SAASKIT_DATABASE_PASSWORD` | PostgreSQL password | — |
| `SAASKIT_DATABASE_NAME` | PostgreSQL database name | `saaskit` |
| `SAASKIT_DATABASE_SSLMODE` | SSL mode (`disable`, `require`, `verify-full`) | `disable` |
| `SAASKIT_JWT_ALGORITHM` | Signing algorithm (`RS256`, `ES256`, `EdDSA`) | `RS256` |
| `SAASKIT_JWT_KEY_PATH` | Path to signing keys | `./keys` |
| `SAASKIT_SERVER_SECRET` | HMAC secret for token hashing | — |
| `SAASKIT_ENCRYPTION_MASTER_KEY` | 32-byte hex key for envelope encryption | — |
| `SAASKIT_OAUTH_GOOGLE_CLIENT_ID` | Google OAuth2 client ID | — |
| `SAASKIT_OAUTH_GOOGLE_CLIENT_SECRET` | Google OAuth2 client secret | — |
| `SAASKIT_OAUTH_GITHUB_CLIENT_ID` | GitHub OAuth2 client ID | — |
| `SAASKIT_OAUTH_GITHUB_CLIENT_SECRET` | GitHub OAuth2 client secret | — |

## Development

```bash
make help              # Show all available commands
make build             # Build binary
make test              # Run tests
make lint              # Run linter
make migrate-up        # Run database migrations
make sqlc              # Regenerate sqlc code
make keys              # Generate development signing keys
make docker-up         # Start with Docker Compose
```

## Technology Stack

| Component | Choice | Why |
|-----------|--------|-----|
| Language | Go 1.22+ | Performance, simplicity, stdlib |
| HTTP | chi/v5 | Lightweight, stdlib-compatible |
| Database | PostgreSQL 16 + pgx/v5 | Gold-standard driver |
| Queries | sqlc | Compile-time type safety |
| Migrations | goose | SQL-first, simple |
| JWT | golang-jwt/v5 | Community standard |
| OIDC | zitadel/oidc/v3 | OIDC-certified |
| Password Hash | Argon2id | OWASP recommended |
| Config | koanf | Lightweight, no global state |
| Logging | slog | Zero dependency, structured |

## Roadmap

See the [full roadmap](docs/roadmap.md) for detailed plans through v3.0.

1. ✅ **Foundation** — scaffold, DB, config, crypto, events, jobs
2. ✅ **Identity** — auth, sessions, OIDC provider, social login, login UI
3. ⬜ **Multi-Tenancy** — organizations, tenant isolation
4. ⬜ **Authorization** — RBAC, permissions middleware
5. ⬜ **Enterprise** — SAML, SCIM, MFA, API keys

## License

[Apache License 2.0](LICENSE)
