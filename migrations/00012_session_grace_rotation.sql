-- +goose Up
ALTER TABLE sessions
    ADD COLUMN previous_refresh_token_hash TEXT,
    ADD COLUMN rotated_at TIMESTAMPTZ,
    ADD COLUMN grace_refresh_token_enc TEXT;

CREATE INDEX idx_sessions_previous_refresh_hash ON sessions (previous_refresh_token_hash)
    WHERE previous_refresh_token_hash IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_sessions_previous_refresh_hash;

ALTER TABLE sessions
    DROP COLUMN IF EXISTS grace_refresh_token_enc,
    DROP COLUMN IF EXISTS rotated_at,
    DROP COLUMN IF EXISTS previous_refresh_token_hash;
