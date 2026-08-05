-- name: CreateAPIKey :one
INSERT INTO api_keys (
    tenant_id,
    name,
    key_prefix,
    key_hash,
    scopes,
    type,
    status,
    expires_at,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetAPIKeyByID :one
SELECT * FROM api_keys
WHERE id = $1 AND tenant_id = $2;

-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys
WHERE key_hash = $1;

-- name: ListAPIKeysByTenant :many
SELECT * FROM api_keys
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: UpdateAPIKeyLastUsed :one
UPDATE api_keys
SET last_used_at = NOW()
WHERE id = $1
RETURNING *;

-- name: RevokeAPIKey :one
UPDATE api_keys
SET status = 'revoked',
    revoked_at = NOW(),
    revoked_by = $2
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: DeleteAPIKey :exec
DELETE FROM api_keys
WHERE id = $1 AND tenant_id = $2;
