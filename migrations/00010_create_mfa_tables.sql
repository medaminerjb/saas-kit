-- +goose Up
-- MFA tables reserved for future implementation.
-- Created now to avoid painful schema migrations when MFA is added.

CREATE TABLE mfa_methods (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type            TEXT NOT NULL,       -- 'totp', 'webauthn', 'sms', 'email', 'recovery_codes'
    name            TEXT,                -- user-friendly label
    secret_enc      TEXT,                -- encrypted TOTP secret or WebAuthn credential
    verified        BOOLEAN NOT NULL DEFAULT FALSE,
    is_default      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at    TIMESTAMPTZ
);

CREATE INDEX idx_mfa_methods_user ON mfa_methods (user_id);

CREATE TABLE mfa_recovery_codes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash       TEXT NOT NULL,
    used_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mfa_recovery_user ON mfa_recovery_codes (user_id);

CREATE TABLE trusted_devices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_hash     TEXT NOT NULL,
    name            TEXT,
    last_used_at    TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trusted_devices_user ON trusted_devices (user_id);

-- +goose Down
DROP TABLE IF EXISTS trusted_devices;
DROP TABLE IF EXISTS mfa_recovery_codes;
DROP TABLE IF EXISTS mfa_methods;
