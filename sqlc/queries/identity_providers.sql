-- name: CreateIdentityProvider :one
INSERT INTO identity_providers (id, tenant_id, name, provider_type, client_id, client_secret_enc, discovery_url, scopes, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetIdentityProviderByID :one
SELECT * FROM identity_providers
WHERE id = $1 AND is_active = true;

-- name: ListIdentityProviders :many
SELECT * FROM identity_providers
WHERE (tenant_id = sqlc.narg('tenant_id') OR sqlc.narg('tenant_id')::uuid IS NULL)
  AND is_active = true
ORDER BY name ASC;

-- name: GetIdentityProviderByType :one
SELECT * FROM identity_providers
WHERE provider_type = $1 AND is_active = true
  AND (tenant_id = sqlc.narg('tenant_id') OR sqlc.narg('tenant_id')::uuid IS NULL)
LIMIT 1;

-- name: CreateExternalAccount :one
INSERT INTO external_accounts (id, user_id, provider_id, external_id, email, profile_data)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetExternalAccountByProviderAndExternalID :one
SELECT * FROM external_accounts
WHERE provider_id = $1 AND external_id = $2;

-- name: GetExternalAccountsByUserID :many
SELECT ea.*, ip.name as provider_name, ip.provider_type
FROM external_accounts ea
JOIN identity_providers ip ON ea.provider_id = ip.id
WHERE ea.user_id = $1;

-- name: DeleteExternalAccount :exec
DELETE FROM external_accounts WHERE id = $1;
