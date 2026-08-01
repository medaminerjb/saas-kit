# ADR-0004: Use zitadel/oidc for OpenID Connect Provider

## Status

Accepted

## Context

SaaSKit acts as both an OpenID Provider (OP) and a Relying Party (RP). We need a library that handles the OIDC protocol correctly. Options considered:

1. **ory/fosite** — Powerful, security-first, but complex and low-level. Powers Ory Hydra.
2. **zitadel/oidc** — OIDC-certified, clean abstractions for both OP and RP, actively maintained.
3. **Custom implementation** — Maximum control but enormous surface area for security bugs.

## Decision

Use **zitadel/oidc v3** for the OpenID Provider implementation.

## Rationale

- **OIDC-certified** — Passes the OpenID Foundation conformance tests.
- **Clean OP abstraction** — The `op.Storage` interface provides a clear contract for what we need to implement.
- **RP support** — Same library handles social login (Relying Party) flows.
- **Active maintenance** — Backed by the Zitadel team, with regular releases.
- **Not an IdP product** — Unlike Zitadel-the-product, the library is a building block, not a pre-built solution.

## Consequences

- We implement the `op.Storage` interface backed by PostgreSQL.
- Login/consent UI is our responsibility (the library is headless).
- We depend on the library's release cycle for security patches.
