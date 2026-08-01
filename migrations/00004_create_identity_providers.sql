-- +goose Up
CREATE TABLE identity_providers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID,
    slug                TEXT NOT NULL UNIQUE,
    name                TEXT NOT NULL,
    protocol            idp_protocol NOT NULL,
    client_id           TEXT,
    client_secret_enc   TEXT,
    issuer_url          TEXT,
    authorization_url   TEXT,
    token_url           TEXT,
    userinfo_url        TEXT,
    scopes              TEXT[] NOT NULL DEFAULT '{openid,profile,email}',
    enabled             BOOLEAN NOT NULL DEFAULT TRUE,
    auto_register       BOOLEAN NOT NULL DEFAULT TRUE,
    config              JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE external_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_id     UUID NOT NULL REFERENCES identity_providers(id),
    external_id     TEXT NOT NULL,
    email           CITEXT,
    name            TEXT,
    avatar_url      TEXT,
    access_token    TEXT,
    refresh_token   TEXT,
    token_expiry    TIMESTAMPTZ,
    raw_data        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider_id, external_id)
);

CREATE INDEX idx_external_accounts_user_id ON external_accounts (user_id);

-- +goose Down
DROP TABLE IF EXISTS external_accounts;
DROP TABLE IF EXISTS identity_providers;
