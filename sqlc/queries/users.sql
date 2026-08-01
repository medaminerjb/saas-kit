-- name: CreateUser :one
INSERT INTO users (email, name, password_hash, status, email_verified, avatar_url, tenant_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
  AND deleted_at IS NULL
  AND (tenant_id = $2 OR (tenant_id IS NULL AND $2::uuid IS NULL));

-- name: UpdateUser :one
UPDATE users
SET name = COALESCE(sqlc.narg('name'), name),
    email = COALESCE(sqlc.narg('email'), email),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    status = COALESCE(sqlc.narg('status'), status),
    email_verified = COALESCE(sqlc.narg('email_verified'), email_verified),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = NOW(), status = 'deleted', updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SetEmailVerified :exec
UPDATE users
SET email_verified = TRUE, status = 'active', updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
  AND (tenant_id = sqlc.narg('tenant_id') OR sqlc.narg('tenant_id')::uuid IS NULL)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE deleted_at IS NULL
  AND (tenant_id = sqlc.narg('tenant_id') OR sqlc.narg('tenant_id')::uuid IS NULL);
