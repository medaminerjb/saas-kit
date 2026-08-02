# SaaSKit Authentication Guide

SaaSKit includes a comprehensive, hardened identity engine that handles user signup, login, session validation, token rotation, and account security.

---

## 1. Authentication Flows

SaaSKit provides stateless API-based authentication using **Asymmetric JWTs** for access validation and **HMAC-hashed refresh tokens** for session tracking.

### Registration Flow
1. Client calls `POST /api/v1/auth/register` with name, email, and password.
2. SaaSKit validates the password strength and email format.
3. The password is hashed using Argon2id.
4. User status is set to `pending_verification`, and a new session is created.
5. SaaSKit fires a `user.registered` event.

### Login Flow
1. Client calls `POST /api/v1/auth/login` with email and password.
2. If credentials match, a session and access/refresh token pair are returned.
3. If password validation fails:
   * The lockout system increments the failure count for the email.
   * On the **5th consecutive failure**, the account transitions to `locked` status, and a `user.locked` event is fired.
   * Lockout blocks subsequent login attempts.
4. If login is successful:
   * The failure count is reset to 0.
   * The system logs the client IP and User-Agent.
   * A `user.login` event is published.

### Token Refresh Flow
To maintain a continuous session without prompting the user for password credentials repeatedly:
1. Client calls `POST /api/v1/auth/refresh` containing the current `refresh_token`.
2. SaaSKit hashes the token and queries the active session in PostgreSQL.
3. On validation:
   * The old session/refresh token is immediately **revoked** (refresh token rotation).
   * A new session and a new pair of access and refresh tokens are issued and returned to the client.
4. If a client attempts to reuse a previously revoked refresh token, the refresh operation fails with `401 Unauthorized`.

### Logout Flow
1. The client makes a request to `POST /api/v1/auth/logout` passing the JWT in the `Authorization: Bearer <JWT>` header.
2. SaaSKit extracts the session ID from the JWT claims.
3. The session is immediately marked as revoked (`revoked_at = NOW()`) in the database.
4. Any subsequent attempt to use the corresponding refresh token fails.

---

## 2. Built-in Security Mitigations

### Account Lockout
* **Failure Limit**: 5 attempts.
* **Mechanism**: In-memory thread-safe tracking within `AuthService`.
* **State Transition**: User status goes from `active`/`pending_verification` to `locked`.
* **Unlock Trigger**: Reset/unlock via administration API or password reset verification.

### IP-Based Rate Limiting
* **Default limits**: 20 requests per second per IP, with a burst threshold of 40 requests.
* **Middleware**: Configured globally in the HTTP router.
* **Response**: Returns `429 Too Many Requests` when limits are exceeded.

### Envelope Encryption
* User OAuth credentials, external tokens, and client secrets are protected at rest using envelope encryption (AES-256-GCM) before being stored in the database.
