-- name: CreateWebhookSubscription :one
INSERT INTO webhook_subscriptions (
    tenant_id,
    url,
    secret,
    event_types,
    status
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetWebhookSubscriptionByID :one
SELECT * FROM webhook_subscriptions
WHERE id = $1 AND tenant_id = $2;

-- name: ListWebhookSubscriptionsByTenant :many
SELECT * FROM webhook_subscriptions
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: ListActiveWebhookSubscriptionsByTenant :many
SELECT * FROM webhook_subscriptions
WHERE tenant_id = $1 AND status = 'active'
ORDER BY created_at DESC;

-- name: UpdateWebhookSubscription :one
UPDATE webhook_subscriptions
SET url = $3,
    secret = $4,
    event_types = $5,
    status = $6,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: DeleteWebhookSubscription :exec
DELETE FROM webhook_subscriptions
WHERE id = $1 AND tenant_id = $2;

-- name: CreateWebhookDelivery :one
INSERT INTO webhook_deliveries (
    subscription_id,
    event_id,
    event_type,
    status_code,
    error_message,
    success
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ListFailedWebhookDeliveries :many
SELECT * FROM webhook_deliveries
WHERE success = false
ORDER BY attempted_at DESC
LIMIT 100;
