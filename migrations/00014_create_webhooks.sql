-- +goose Up
CREATE TABLE webhook_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    secret TEXT,
    event_types TEXT[] NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ
);

CREATE INDEX idx_webhook_subscriptions_tenant_id ON webhook_subscriptions(tenant_id);
CREATE INDEX idx_webhook_subscriptions_status ON webhook_subscriptions(status) WHERE status = 'active';

CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES webhook_subscriptions(id) ON DELETE CASCADE,
    event_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    status_code INTEGER,
    error_message TEXT,
    success BOOLEAN NOT NULL DEFAULT false,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_deliveries_subscription_id ON webhook_deliveries(subscription_id);
CREATE INDEX idx_webhook_deliveries_event_id ON webhook_deliveries(event_id);
CREATE INDEX idx_webhook_deliveries_success ON webhook_deliveries(success) WHERE success = false;

-- +goose Down
DROP INDEX IF EXISTS idx_webhook_deliveries_success;
DROP INDEX IF EXISTS idx_webhook_deliveries_event_id;
DROP INDEX IF EXISTS idx_webhook_deliveries_subscription_id;
DROP TABLE IF EXISTS webhook_deliveries;
DROP INDEX IF EXISTS idx_webhook_subscriptions_status;
DROP INDEX IF EXISTS idx_webhook_subscriptions_tenant_id;
DROP TABLE IF EXISTS webhook_subscriptions;
