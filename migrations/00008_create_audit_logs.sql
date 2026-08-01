-- +goose Up
CREATE TABLE audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID,
    actor_id    UUID,
    target_id   UUID,
    event       audit_event_type NOT NULL,
    ip_address  INET,
    user_agent  TEXT,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_tenant ON audit_logs (tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX idx_audit_logs_actor ON audit_logs (actor_id) WHERE actor_id IS NOT NULL;
CREATE INDEX idx_audit_logs_event ON audit_logs (event);
CREATE INDEX idx_audit_logs_created ON audit_logs (created_at);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
