# SaaSKit Configuration Guide

SaaSKit uses `koanf` for dynamic configuration management. Configuration can be defined using YAML files and overridden using environment variables prefixed with `SAASKIT_`. Environment variables always take precedence over file-based configuration.

---

## 1. Config Loading Order

1. **Defaults**: Hardcoded fallback values within the application.
2. **YAML Config File**: Loaded from `config.yaml` or a path specified by the `SAASKIT_CONFIG_PATH` env variable.
3. **Environment Variables**: Prefixed with `SAASKIT_` (e.g., `SAASKIT_DATABASE_HOST` maps to `Database.Host`).

---

## 2. Configuration Parameters Reference

Below is a complete reference of the configuration keys available in SaaSKit v0.1.0:

### Server Configuration

| YAML Key | Environment Variable | Type | Default | Description |
|---|---|---|---|---|
| `env` | `SAASKIT_ENV` | `string` | `development` | App environment: `development` or `production`. |
| `port` | `SAASKIT_PORT` | `string` | `8080` | Port the HTTP server will bind to. |
| `base_url` | `SAASKIT_BASE_URL` | `string` | `http://localhost:8080` | Public URL of the server (used for tokens/OAuth redirects). |
| `server_secret` | `SAASKIT_SERVER_SECRET` | `string` | - | Server HMAC key used for hashing refresh and reset tokens. |
| `encryption_master_key` | `SAASKIT_ENCRYPTION_MASTER_KEY` | `string` | - | Hex-encoded 32-byte master key for envelope database encryption. |

### Database Configuration

| YAML Key | Environment Variable | Type | Default | Description |
|---|---|---|---|---|
| `database.host` | `SAASKIT_DATABASE_HOST` | `string` | `localhost` | PostgreSQL host. |
| `database.port` | `SAASKIT_DATABASE_PORT` | `string` | `5432` | PostgreSQL port. |
| `database.user` | `SAASKIT_DATABASE_USER` | `string` | `postgres` | PostgreSQL username. |
| `database.password` | `SAASKIT_DATABASE_PASSWORD` | `string` | `postgres` | PostgreSQL password. |
| `database.name` | `SAASKIT_DATABASE_NAME` | `string` | `saaskit` | PostgreSQL database name. |
| `database.sslmode` | `SAASKIT_DATABASE_SSLMODE` | `string` | `disable` | PostgreSQL SSL mode (`disable`, `require`, `verify-full`). |
| `database.max_conns` | `SAASKIT_DATABASE_MAX_CONNS` | `int` | `20` | Maximum open database connections in pgxpool. |
| `database.min_conns` | `SAASKIT_DATABASE_MIN_CONNS` | `int` | `2` | Minimum idle database connections. |

### JWT Key Storage Configuration

| YAML Key | Environment Variable | Type | Default | Description |
|---|---|---|---|---|
| `jwt.algorithm` | `SAASKIT_JWT_ALGORITHM` | `string` | `RS256` | Token signing algorithm (`RS256`, `ES256`, `EdDSA`). |
| `jwt.key_path` | `SAASKIT_JWT_KEY_PATH` | `string` | `./keys` | Directory path containing private/public keys. |
| `jwt.issuer` | `SAASKIT_JWT_ISSUER` | `string` | `http://localhost:8080` | Expected token issuer (`iss` claim). |
| `jwt.access_ttl` | `SAASKIT_JWT_ACCESS_TTL` | `duration` | `15m` | Lifetime of JWT access tokens. |
| `jwt.refresh_ttl` | `SAASKIT_JWT_REFRESH_TTL` | `duration` | `168h` | Lifetime of refresh tokens (7 days). |

### Logging Configuration

| YAML Key | Environment Variable | Type | Default | Description |
|---|---|---|---|---|
| `log.level` | `SAASKIT_LOG_LEVEL` | `string` | `info` | Minimum log severity: `debug`, `info`, `warn`, `error`. |
| `log.format` | `SAASKIT_LOG_FORMAT` | `string` | `json` | Logging format: `json` or `text`. |

---

## 3. Production Configuration Best Practices

> [!WARNING]
> Never use default keys, credentials, or simple development configurations in production.

* **Envelope Encryption**: Ensure your `SAASKIT_ENCRYPTION_MASTER_KEY` is exactly a 32-byte hex-encoded string (64 characters). This key is used to encrypt sensitive columns (like client credentials) in PostgreSQL.
* **Server Secrets**: Keep `SAASKIT_SERVER_SECRET` long and random.
* **Asymmetric Keys**: In production, generate signing keys securely and mount them to the container path configured via `SAASKIT_JWT_KEY_PATH`. Make sure the private key files are read-only (`chmod 400`).
* **Database Connection Security**: Always set `SAASKIT_DATABASE_SSLMODE` to `require` or `verify-full` when connecting to a managed cloud database.
