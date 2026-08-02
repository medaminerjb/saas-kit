-- name: GetOIDCClientByID :one
SELECT id, tenant_id, client_name, client_secret_hash, redirect_uris, 
       post_logout_redirect_uris, response_types, grant_types,
       scopes, token_endpoint_auth_method, pkce_required,
       consent_required, application_type, jwks_uri, jwks,
       logo_uri, tos_uri, policy_uri, contacts,
       access_token_ttl, refresh_token_ttl, id_token_ttl,
       is_public, disabled, created_at, updated_at
FROM oidc_clients
WHERE id = $1 AND disabled = false;

-- name: CreateOIDCClient :one
INSERT INTO oidc_clients (
    id, tenant_id, client_name, client_secret_hash,
    redirect_uris, post_logout_redirect_uris,
    response_types, grant_types, scopes,
    token_endpoint_auth_method, pkce_required,
    consent_required, application_type,
    access_token_ttl, refresh_token_ttl, id_token_ttl,
    is_public, disabled
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
RETURNING *;

-- name: ListOIDCClients :many
SELECT * FROM oidc_clients
WHERE (tenant_id = sqlc.narg('tenant_id') OR sqlc.narg('tenant_id')::uuid IS NULL)
  AND disabled = false
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateOIDCClient :one
UPDATE oidc_clients
SET client_name = COALESCE(sqlc.narg('client_name'), client_name),
    redirect_uris = COALESCE(sqlc.narg('redirect_uris'), redirect_uris),
    post_logout_redirect_uris = COALESCE(sqlc.narg('post_logout_redirect_uris'), post_logout_redirect_uris),
    disabled = COALESCE(sqlc.narg('disabled'), disabled),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteOIDCClient :exec
UPDATE oidc_clients SET disabled = true, updated_at = NOW()
WHERE id = $1;
