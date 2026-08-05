-- +goose Up
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key_prefix TEXT NOT NULL, -- sk_live_ or sk_test_
    key_hash TEXT NOT NULL, -- SHA-256 hash of the full key
    scopes TEXT[] NOT NULL DEFAULT '{}',
    type TEXT NOT NULL CHECK (type IN ('test', 'live')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked', 'expired')),
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    revoked_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_key_status ON api_keys(status) WHERE status = 'active';
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at) WHERE expires_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_api_keys_expires_at;
DROP INDEX IF EXISTS idx_api_key_status;
DROP INDEX IF EXISTS idx_api_keys_key_hash;
DROP INDEX IF EXISTS idx_api_keys_tenant_id;
DROP TABLE IF EXISTS api_keys;
