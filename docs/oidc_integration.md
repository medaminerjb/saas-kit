# SaaSKit OIDC Integration Guide

SaaSKit includes an OIDC (OpenID Connect) and OAuth 2.0 provider powered by `zitadel/oidc/v3`. It acts as an identity provider, allowing client applications to authenticate users using standardized protocols.

---

## 1. Provider Endpoints

SaaSKit exposes standard OIDC discovery and authorization endpoints:

* **OpenID Provider Discovery**: `GET /.well-known/openid-configuration`
* **JSON Web Key Set (JWKS)**: `GET /.well-known/jwks.json`
* **Authorization Endpoint**: `GET/POST /oauth/v2/authorize`
* **Token Endpoint**: `POST /oauth/v2/token`
* **Userinfo Endpoint**: `GET/POST /oauth/v2/userinfo`
* **End Session (Logout) Endpoint**: `GET/POST /oauth/v2/end_session`

---

## 2. Supported Authorization Flows

SaaSKit supports standard OAuth 2.0 and OpenID Connect flows, focusing on secure, modern practices:

### Authorization Code Flow with PKCE
* Recommended for single-page apps (SPAs), mobile apps, and server-side web apps.
* Clients must provide `code_challenge` and `code_challenge_method` (`S256`) during the authorization request.
* Clients must verify using `code_verifier` during the token request.

### Client Credentials Flow
* Supported for machine-to-machine (M2M) backend services.

---

## 3. Client Registration

To authorize applications, OIDC clients must be registered in the database.

### Schema Table: `oidc_clients`
Clients are registered by inserting records into the `oidc_clients` table:

| Column | Type | Description |
|---|---|---|
| `id` | `TEXT` | Unique client identifier (e.g. `my-web-app`). |
| `client_secret_enc` | `TEXT` | AES-256-GCM encrypted client secret (for confidential clients). |
| `redirect_uris` | `TEXT[]` | Allowed list of redirect URIs. |
| `response_types` | `TEXT[]` | Supported response types (e.g. `code`). |
| `grant_types` | `TEXT[]` | Supported grant types (e.g. `authorization_code`, `refresh_token`). |
| `application_type` | `TEXT` | `web` or `native`. |
| `auth_method` | `TEXT` | Client authentication method (e.g., `client_secret_post`, `none` for PKCE). |
| `contacts` | `TEXT[]` | List of email contacts for the client. |

### Example Client Insertion
Below is an SQL statement to register a confidential web application using the DB:
```sql
INSERT INTO oidc_clients (
    id,
    client_secret_enc,
    redirect_uris,
    response_types,
    grant_types,
    application_type,
    auth_method
) VALUES (
    'my-spa-client',
    NULL, -- Public client using PKCE has no secret
    ARRAY['http://localhost:3000/callback'],
    ARRAY['code'],
    ARRAY['authorization_code', 'refresh_token'],
    'native',
    'none'
);
```

---

## 4. Integration Example

For an OIDC-compliant client (e.g., using `oidc-client-ts`, NextAuth, or Go's `golang.org/x/oauth2`), configure the following settings:

* **Authority / Issuer**: `http://localhost:8080` (or your configured public base URL)
* **Client ID**: `my-spa-client`
* **Redirect URI**: `http://localhost:3000/callback`
* **Scopes**: `openid profile email offline_access`
