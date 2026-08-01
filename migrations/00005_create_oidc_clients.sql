-- +goose Up
CREATE TABLE oidc_clients (
    id                          TEXT PRIMARY KEY,
    tenant_id                   UUID,
    client_name                 TEXT NOT NULL,
    client_secret_hash          TEXT,
    redirect_uris               TEXT[] NOT NULL,
    post_logout_redirect_uris   TEXT[],
    grant_types                 TEXT[] NOT NULL DEFAULT '{authorization_code}',
    response_types              TEXT[] NOT NULL DEFAULT '{code}',
    scopes                      TEXT[] NOT NULL DEFAULT '{openid,profile,email}',
    token_endpoint_auth_method  TEXT NOT NULL DEFAULT 'client_secret_basic',
    pkce_required               BOOLEAN NOT NULL DEFAULT TRUE,
    consent_required            BOOLEAN NOT NULL DEFAULT TRUE,
    application_type            TEXT NOT NULL DEFAULT 'web',
    jwks_uri                    TEXT,
    jwks                        JSONB,
    logo_uri                    TEXT,
    tos_uri                     TEXT,
    policy_uri                  TEXT,
    contacts                    TEXT[],
    access_token_ttl            INTERVAL DEFAULT '15 minutes',
    refresh_token_ttl           INTERVAL DEFAULT '7 days',
    id_token_ttl                INTERVAL DEFAULT '1 hour',
    is_public                   BOOLEAN NOT NULL DEFAULT FALSE,
    disabled                    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS oidc_clients;
