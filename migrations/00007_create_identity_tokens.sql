-- +goose Up
CREATE TABLE identity_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        token_type NOT NULL,
    hash        TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_identity_tokens_hash ON identity_tokens (hash);
CREATE INDEX idx_identity_tokens_user_type ON identity_tokens (user_id, type);
CREATE INDEX idx_identity_tokens_expires ON identity_tokens (expires_at) WHERE used_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS identity_tokens;
