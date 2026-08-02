# SaaSKit — Architecture Specification

This document defines the core architecture, design decisions, and data flows for SaaSKit. It serves as the authoritative guide for maintainers, contributors, and operators.

---

## 1. Design Philosophy

SaaSKit is built on three core pillars:
1. **Tenancy-First:** Multi-tenancy is not an afterthought. Every database table, config schema, and routing flow is designed to support isolation out of the box.
2. **Standard-Compliant:** We rely on industry standards (OIDC, OAuth2, JWT, Argon2id, AES-GCM) rather than inventing proprietary security mechanisms.
3. **Modular Expansion:** The core identity and access layer is kept lean. Higher-level features (billing, analytics, webhooks) are implemented as opt-in modules that plug into defined core interfaces.

---

## 2. Structural Layering

SaaSKit uses a clean, decoupled Go service architecture. Code is partitioned to separate business logic from delivery mechanisms and databases.

```text
┌─────────────────────────────────────────────────────────┐
│                    HTTP Router & Handlers               │ (internal/identity/handler)
└────────────────────────────┬────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────┐
│                    IdentityManager                      │ (internal/identity/service)
└───────┬────────────────────┬────────────────────┬───────┘
        ▼                    ▼                    ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  AuthService │     │  UserService │     │ TokenService │ (internal/identity/service)
└───────┬──────┘     └───────┬──────┘     └───────┬──────┘
        ▼                    ▼                    ▼
┌─────────────────────────────────────────────────────────┐
│                  Repository Interfaces                  │ (internal/identity/repository)
└────────────────────────────┬────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────┐
│                   PostgreSQL (pgx, sqlc)                │ (internal/identity/repository)
└─────────────────────────────────────────────────────────┘
```

### Components
* **Domain Layer (`internal/identity/domain/`):** Contains the pure business entities (e.g., `User`, `Session`, `IdentityToken`) and domain errors (e.g., `ErrUserNotFound`).
* **Service Layer (`internal/identity/service/`):** Contains the core business logic (e.g., checking passwords, generating sessions, issuing JWTs).
* **IdentityManager:** Coordinates operations across multiple sub-services (e.g., triggering registration, MFA checks, audit events, or welcome emails).
* **Repository Layer (`internal/identity/repository/`):** Implements data access using PostgreSQL (`pgx` and `sqlc`).
* **HTTP Layer (`internal/identity/handler/`):** Handles serialization, input validation, and routing.

---

## 3. Core Modules

### 3.1. Authentication & User Management
SaaSKit uses **Argon2id** (configured according to OWASP guidelines) for password hashing. The user model implements a 6-state lifecycle:
- `pending`: Registered, awaiting email verification.
- `active`: Fully operational.
- `disabled`: Administrative suspension.
- `locked`: Locked due to failed login threshold.
- `invited`: Invited to join a tenant, account incomplete.
- `deleted`: Soft-deleted (retains unique indexes but inaccessible via queries).

### 3.2. Session & Token Management
* **Refresh Token Rotation:** Refresh tokens are rotated on every write. When a client requests a new access token using a refresh token, SaaSKit revokes the old session and issues a new session with a new refresh token.
* **Token Hashing:** Raw refresh tokens are never persisted in the database. Instead, they are hashed using HMAC-SHA256 with the server's master key. Even in the event of a full database leak, attackers cannot hijack active sessions without the server key.

### 3.3. OpenID Connect (OIDC) Provider
SaaSKit includes a certified OIDC Provider implemented via `zitadel/oidc/v3`. It bridges our database repositories to the standard OIDC code flow:
- Supports **PKCE** (Proof Key for Code Exchange) out of the box.
- Serves OIDC discovery metadata at `/.well-known/openid-configuration`.
- Manages signing keys dynamically or loads static PEM credentials from disk.

### 3.4. Cryptographic Infrastructure
* **Envelope Encryption:** Sensitive database values (OAuth client secrets, external provider tokens, MFA backup secrets) are encrypted before write. SaaSKit uses AES-256-GCM envelope encryption with a master key derived via HKDF-SHA256, preventing data leaks from database dumps.

---

## 4. Telemetry and Audit Trail

All identity actions publish audit events containing:
- `ActorID`: Who initiated the change.
- `TargetID`: The resource affected.
- `Type`: Event name (e.g., `user.registered`, `user.login`).
- `Metadata`: Extra context (IP address, user agent).

Events are handled by the `events.Publisher` interface, enabling simple log-based output (`slog`) or zero-dependency extension to message queues (Kafka, NATS, RabbitMQ).
