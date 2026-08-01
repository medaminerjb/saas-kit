# Signing Keys

This directory holds the JWT signing keys used by SaaSKit.

## Development

In development mode (`SAASKIT_ENV=development`), keys are auto-generated on first boot if missing.

Generate manually:

```bash
make keys
```

## Production

**Keys must be provisioned manually.** SaaSKit will **refuse to start** in production if signing keys are missing.

### RS256 (default)

```bash
openssl genpkey -algorithm RSA -out keys/active.pem -pkeyopt rsa_keygen_bits:4096
openssl rsa -pubout -in keys/active.pem -out keys/active.pub
```

### ES256

```bash
openssl ecparam -genkey -name prime256v1 -noout -out keys/active.pem
openssl ec -in keys/active.pem -pubout -out keys/active.pub
```

### EdDSA (Ed25519)

```bash
openssl genpkey -algorithm Ed25519 -out keys/active.pem
openssl pkey -in keys/active.pem -pubout -out keys/active.pub
```

## Key Rotation

To rotate keys:

1. Rename `active.pem` → `previous.pem` (and `.pub`)
2. Generate new `active.pem` + `active.pub`
3. Restart SaaSKit

Both keys will be served in the JWKS endpoint (`/oidc/keys`) during the transition.
Old tokens signed with `previous` remain valid until they expire.

## Security

- **Never commit `.pem` or `.pub` files to version control.**
- The `.gitignore` is configured to exclude them.
- In production, use a secrets manager (Vault, AWS Secrets Manager, etc.).
