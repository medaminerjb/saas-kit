-- name: GetOIDCClientByID :one
SELECT id, tenant_id, client_name, client_secret_hash, redirect_uris, 
       post_logout_redirect_uris, response_types, grant_types,
       application_type, token_endpoint_auth_method, pkce_required,
       scope_restrictions, id_token_lifetime, access_token_lifetime,
       refresh_token_lifetime, access_token_type, logo_uri, client_uri,
       policy_uri, tos_uri, is_active
FROM oidc_clients
WHERE id = $1 AND is_active = true;

-- name: CreateOIDCClient :one
INSERT INTO oidc_clients (
    id, tenant_id, client_name, client_secret_hash,
    redirect_uris, post_logout_redirect_uris,
    response_types, grant_types, application_type,
    token_endpoint_auth_method, pkce_required,
    scope_restrictions, access_token_type
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: ListOIDCClients :many
SELECT * FROM oidc_clients
WHERE (tenant_id = sqlc.narg('tenant_id') OR sqlc.narg('tenant_id')::uuid IS NULL)
  AND is_active = true
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateOIDCClient :one
UPDATE oidc_clients
SET client_name = COALESCE(sqlc.narg('client_name'), client_name),
    redirect_uris = COALESCE(sqlc.narg('redirect_uris'), redirect_uris),
    post_logout_redirect_uris = COALESCE(sqlc.narg('post_logout_redirect_uris'), post_logout_redirect_uris),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteOIDCClient :exec
UPDATE oidc_clients SET is_active = false, updated_at = NOW()
WHERE id = $1;
