-- name: CreateMFAMethod :one
INSERT INTO mfa_methods (user_id, type, name, secret_enc, verified, is_default)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetMFAMethodByID :one
SELECT * FROM mfa_methods WHERE id = $1 AND user_id = $2;

-- name: ListMFAMethodsForUser :many
SELECT * FROM mfa_methods WHERE user_id = $1 ORDER BY created_at;

-- name: GetVerifiedTOTPForUser :one
SELECT * FROM mfa_methods
WHERE user_id = $1 AND type = 'totp' AND verified = TRUE
LIMIT 1;

-- name: HasVerifiedMFAForUser :one
SELECT EXISTS(
    SELECT 1 FROM mfa_methods WHERE user_id = $1 AND verified = TRUE
) AS has_mfa;

-- name: MarkMFAMethodVerified :one
UPDATE mfa_methods SET verified = TRUE
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteMFAMethod :exec
DELETE FROM mfa_methods WHERE id = $1 AND user_id = $2;

-- name: DeleteAllMFAMethodsForUser :exec
DELETE FROM mfa_methods WHERE user_id = $1;

-- name: CreateMFARecoveryCode :exec
INSERT INTO mfa_recovery_codes (user_id, code_hash)
VALUES ($1, $2);

-- name: ListUnusedRecoveryCodes :many
SELECT * FROM mfa_recovery_codes
WHERE user_id = $1 AND used_at IS NULL;

-- name: MarkRecoveryCodeUsed :exec
UPDATE mfa_recovery_codes SET used_at = NOW()
WHERE id = $1 AND user_id = $2 AND used_at IS NULL;

-- name: DeleteRecoveryCodesForUser :exec
DELETE FROM mfa_recovery_codes WHERE user_id = $1;

-- name: CountRecoveryCodesForUser :one
SELECT COUNT(*) FROM mfa_recovery_codes
WHERE user_id = $1 AND used_at IS NULL;
