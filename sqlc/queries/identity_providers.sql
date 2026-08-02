-- name: CreateIdentityProvider :one
INSERT INTO identity_providers (id, tenant_id, slug, name, protocol, client_id, client_secret_enc, issuer_url, authorization_url, token_url, userinfo_url, scopes, enabled, auto_register, config)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: GetIdentityProviderByID :one
SELECT * FROM identity_providers
WHERE id = $1 AND enabled = true;

-- name: ListIdentityProviders :many
SELECT * FROM identity_providers
WHERE (tenant_id = sqlc.narg('tenant_id') OR sqlc.narg('tenant_id')::uuid IS NULL)
  AND enabled = true
ORDER BY name ASC;

-- name: GetIdentityProviderByProtocol :one
SELECT * FROM identity_providers
WHERE protocol = $1 AND enabled = true
  AND (tenant_id = sqlc.narg('tenant_id') OR sqlc.narg('tenant_id')::uuid IS NULL)
LIMIT 1;

-- name: CreateExternalAccount :one
INSERT INTO external_accounts (id, user_id, provider_id, external_id, email, name, avatar_url, access_token, refresh_token, token_expiry, raw_data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetExternalAccountByProviderAndExternalID :one
SELECT * FROM external_accounts
WHERE provider_id = $1 AND external_id = $2;

-- name: GetExternalAccountsByUserID :many
SELECT ea.*, ip.name as provider_name, ip.protocol as provider_protocol
FROM external_accounts ea
JOIN identity_providers ip ON ea.provider_id = ip.id
WHERE ea.user_id = $1;

-- name: DeleteExternalAccount :exec
DELETE FROM external_accounts WHERE id = $1;
