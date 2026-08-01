-- +goose Up
CREATE TABLE users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID,
    email             CITEXT NOT NULL,
    name              TEXT NOT NULL DEFAULT '',
    password_hash     TEXT,
    status            user_status NOT NULL DEFAULT 'pending_verification',
    email_verified    BOOLEAN NOT NULL DEFAULT FALSE,
    avatar_url        TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at     TIMESTAMPTZ,
    deleted_at        TIMESTAMPTZ
);

-- Scoped uniqueness: email is unique per tenant (or globally when tenant_id IS NULL)
CREATE UNIQUE INDEX idx_users_email_tenant ON users (email, tenant_id) WHERE deleted_at IS NULL AND tenant_id IS NOT NULL;
CREATE UNIQUE INDEX idx_users_email_global ON users (email) WHERE deleted_at IS NULL AND tenant_id IS NULL;
CREATE INDEX idx_users_tenant_id ON users (tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX idx_users_status ON users (status);
CREATE INDEX idx_users_not_deleted ON users (id) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS users;
