-- +goose Up
CREATE TABLE authorization_requests (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id               TEXT NOT NULL REFERENCES oidc_clients(id),
    user_id                 UUID REFERENCES users(id),
    scopes                  TEXT[] NOT NULL,
    redirect_uri            TEXT NOT NULL,
    state                   TEXT,
    nonce                   TEXT,
    code_challenge          TEXT,
    code_challenge_method   TEXT,
    response_type           TEXT NOT NULL,
    authenticated           BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at              TIMESTAMPTZ NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_requests_expires ON authorization_requests (expires_at);

CREATE TABLE authorization_codes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code_hash       TEXT NOT NULL UNIQUE,
    request_id      UUID NOT NULL REFERENCES authorization_requests(id) ON DELETE CASCADE,
    expires_at      TIMESTAMPTZ NOT NULL,
    redeemed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_codes_hash ON authorization_codes (code_hash);
CREATE INDEX idx_auth_codes_expires ON authorization_codes (expires_at) WHERE redeemed_at IS NULL;

CREATE TABLE consents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id       TEXT NOT NULL REFERENCES oidc_clients(id),
    scopes          TEXT[] NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at      TIMESTAMPTZ,
    UNIQUE (user_id, client_id)
);

CREATE TABLE device_codes (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id           TEXT NOT NULL REFERENCES oidc_clients(id),
    device_code_hash    TEXT NOT NULL UNIQUE,
    user_code           TEXT NOT NULL UNIQUE,
    user_id             UUID REFERENCES users(id),
    scopes              TEXT[] NOT NULL,
    verified            BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS device_codes;
DROP TABLE IF EXISTS consents;
DROP TABLE IF EXISTS authorization_codes;
DROP TABLE IF EXISTS authorization_requests;
