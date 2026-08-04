-- name: CreateTenant :one
INSERT INTO tenants (
    name, slug, status, plan, metadata, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, NOW(), NOW()
) RETURNING *;

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: UpdateTenant :one
UPDATE tenants
SET name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    status = COALESCE(sqlc.narg('status'), status),
    plan = COALESCE(sqlc.narg('plan'), plan),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateTenantMetadata :one
UPDATE tenants
SET metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTenant :exec
DELETE FROM tenants WHERE id = $1;

-- name: ListTenantsForUser :many
SELECT t.*, tm.role, tm.joined_at
FROM tenants t
JOIN tenant_members tm ON t.id = tm.tenant_id
WHERE tm.user_id = $1
ORDER BY t.name ASC;

-- name: AddTenantMember :one
INSERT INTO tenant_members (
    tenant_id, user_id, role, joined_at
) VALUES (
    $1, $2, $3, NOW()
) RETURNING *;

-- name: GetTenantMember :one
SELECT * FROM tenant_members
WHERE tenant_id = $1 AND user_id = $2;

-- name: ListTenantMembers :many
SELECT tm.*, u.email, u.name, u.avatar_url
FROM tenant_members tm
JOIN users u ON tm.user_id = u.id
WHERE tm.tenant_id = $1 AND u.deleted_at IS NULL
ORDER BY u.name ASC;

-- name: RemoveTenantMember :exec
DELETE FROM tenant_members
WHERE tenant_id = $1 AND user_id = $2;

-- name: UpdateTenantMemberRole :one
UPDATE tenant_members
SET role = $3
WHERE tenant_id = $1 AND user_id = $2
RETURNING *;

-- name: CreateTenantInvitation :one
INSERT INTO tenant_invitations (
    tenant_id, email, role, token_hash, expires_at, created_at
) VALUES (
    $1, $2, $3, $4, $5, NOW()
) RETURNING *;

-- name: GetTenantInvitationByID :one
SELECT * FROM tenant_invitations WHERE id = $1;

-- name: GetTenantInvitationByToken :one
SELECT * FROM tenant_invitations
WHERE token_hash = $1 AND accepted_at IS NULL AND expires_at > NOW();

-- name: ListTenantInvitations :many
SELECT * FROM tenant_invitations
WHERE tenant_id = $1 AND accepted_at IS NULL AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: AcceptTenantInvitation :one
UPDATE tenant_invitations
SET accepted_at = NOW()
WHERE id = $1 AND accepted_at IS NULL AND expires_at > NOW()
RETURNING *;

-- name: DeleteTenantInvitation :exec
DELETE FROM tenant_invitations WHERE id = $1;

-- name: UpdateUserActiveTenant :one
UPDATE users
SET tenant_id = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
