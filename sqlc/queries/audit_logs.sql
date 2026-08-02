-- name: CreateAuditLog :exec
INSERT INTO audit_logs (tenant_id, actor_id, target_id, event, ip_address, user_agent, metadata)
VALUES (
  sqlc.narg('tenant_id'),
  sqlc.narg('actor_id'),
  sqlc.narg('target_id'),
  sqlc.arg('event'),
  sqlc.narg('ip_address')::inet,
  sqlc.narg('user_agent'),
  sqlc.narg('metadata')
);

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
WHERE (tenant_id = sqlc.narg('tenant_id') OR sqlc.narg('tenant_id')::uuid IS NULL)
  AND (actor_id = sqlc.narg('actor_id') OR sqlc.narg('actor_id')::uuid IS NULL)
  AND (event = sqlc.narg('event') OR sqlc.narg('event')::audit_event_type IS NULL)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAuditLogs :one
SELECT COUNT(*) FROM audit_logs
WHERE (tenant_id = sqlc.narg('tenant_id') OR sqlc.narg('tenant_id')::uuid IS NULL)
  AND (actor_id = sqlc.narg('actor_id') OR sqlc.narg('actor_id')::uuid IS NULL)
  AND (event = sqlc.narg('event') OR sqlc.narg('event')::audit_event_type IS NULL);
