-- name: CreateIdentityToken :one
INSERT INTO identity_tokens (user_id, type, hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetIdentityTokenByHash :one
SELECT * FROM identity_tokens
WHERE hash = $1
  AND used_at IS NULL
  AND expires_at > NOW();

-- name: MarkIdentityTokenUsed :exec
UPDATE identity_tokens SET used_at = NOW() WHERE id = $1;

-- name: InvalidateUserTokensByType :exec
UPDATE identity_tokens SET used_at = NOW()
WHERE user_id = $1 AND type = $2 AND used_at IS NULL;

-- name: DeleteExpiredIdentityTokens :execrows
DELETE FROM identity_tokens
WHERE expires_at < NOW() OR used_at IS NOT NULL;
