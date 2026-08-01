# ADR-0002: Use Asymmetric JWT Signing (RS256/ES256/EdDSA)

## Status

Accepted

## Context

JWTs need to be signed. The two main approaches are:

1. **Symmetric (HS256)** — Single shared secret. Simple but the secret must be shared with every service that needs to verify tokens.
2. **Asymmetric (RS256/ES256/EdDSA)** — Private key signs, public key verifies. Only the issuer holds the private key.

## Decision

Use **asymmetric signing** exclusively. Support RS256 (default), ES256, and EdDSA (Ed25519). HS256 is not supported outside of test fixtures.

## Rationale

- **Zero-trust verification** — Any service can verify tokens using the public JWKS endpoint without needing access to signing secrets.
- **Key rotation** — Multiple public keys can be served simultaneously via JWKS, enabling zero-downtime key rotation.
- **OIDC compliance** — The OpenID Connect specification requires asymmetric signing for interoperability.
- **Security posture** — If a verifying service is compromised, the attacker cannot forge new tokens.

## Consequences

- Operators must provision signing keys in production (SaaSKit refuses to start without them).
- Development mode auto-generates keys for convenience.
- JWKS endpoint (`/oidc/keys`) must serve all active public keys.
- Slightly higher CPU cost for RSA operations (mitigated by ES256/EdDSA options).
