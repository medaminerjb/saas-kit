// Package repository provides PostgreSQL implementations of identity repository interfaces.
package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medaminerjb/saas-kit/internal/identity/domain"
)

// UserRepo implements repository.UserRepository using pgx.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo creates a new PostgreSQL user repository.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// Create inserts a new user.
func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	metadataPublicBytes, err := json.Marshal(user.MetadataPublic)
	if err != nil {
		return err
	}
	metadataPrivateBytes, err := json.Marshal(user.MetadataPrivate)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO users (id, tenant_id, email, name, password_hash, status, email_verified, avatar_url, metadata_public, metadata_private, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err = r.pool.Exec(ctx, query,
		user.ID, user.TenantID, user.Email, user.Name,
		user.PasswordHash, user.Status, user.EmailVerified,
		user.AvatarURL, metadataPublicBytes, metadataPrivateBytes,
		user.CreatedAt, user.UpdatedAt,
	)
	return err
}

// GetByID retrieves a user by ID, excluding soft-deleted users.
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, tenant_id, email, name, password_hash, status, email_verified, avatar_url, metadata_public, metadata_private, created_at, updated_at, last_login_at, deleted_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`

	return r.scanUser(r.pool.QueryRow(ctx, query, id))
}

// GetByEmail retrieves a user by email, scoped to an optional tenant.
func (r *UserRepo) GetByEmail(ctx context.Context, email string, tenantID *uuid.UUID) (*domain.User, error) {
	query := `SELECT id, tenant_id, email, name, password_hash, status, email_verified, avatar_url, metadata_public, metadata_private, created_at, updated_at, last_login_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
		AND (tenant_id = $2 OR (tenant_id IS NULL AND $2::uuid IS NULL))`

	return r.scanUser(r.pool.QueryRow(ctx, query, email, tenantID))
}

// Update updates a user's mutable fields.
func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET name = $2, email = $3, avatar_url = $4, status = $5, email_verified = $6, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.pool.Exec(ctx, query, user.ID, user.Name, user.Email, user.AvatarURL, user.Status, user.EmailVerified)
	return err
}

// UpdateMetadata updates a user's metadata fields.
func (r *UserRepo) UpdateMetadata(ctx context.Context, id uuid.UUID, metadataPublic, metadataPrivate map[string]interface{}) error {
	query := `UPDATE users SET metadata_public = $2, metadata_private = $3, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.pool.Exec(ctx, query, id, metadataPublic, metadataPrivate)
	return err
}

// UpdatePassword updates a user's password hash.
func (r *UserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, id, passwordHash)
	return err
}

// UpdateLastLogin sets the user's last login timestamp.
func (r *UserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET last_login_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// SetEmailVerified marks a user's email as verified and activates the account.
func (r *UserRepo) SetEmailVerified(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET email_verified = TRUE, status = 'active', updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// SoftDelete performs a soft delete on a user.
func (r *UserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW(), status = 'deleted', updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// List returns a paginated list of users.
func (r *UserRepo) List(ctx context.Context, tenantID *uuid.UUID, limit, offset int32) ([]*domain.User, error) {
	query := `SELECT id, tenant_id, email, name, password_hash, status, email_verified, avatar_url, metadata_public, metadata_private, created_at, updated_at, last_login_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL AND (tenant_id = $1 OR $1::uuid IS NULL)
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u, err := r.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// Count returns the total number of users.
func (r *UserRepo) Count(ctx context.Context, tenantID *uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND (tenant_id = $1 OR $1::uuid IS NULL)`
	var count int64
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(&count)
	return count, err
}

// scanUser is a pgx.Row or pgx.Rows scanner for domain.User.
type scannable interface {
	Scan(dest ...any) error
}

func (r *UserRepo) scanUser(row scannable) (*domain.User, error) {
	var u domain.User
	var metadataPublicBytes, metadataPrivateBytes []byte
	err := row.Scan(
		&u.ID, &u.TenantID, &u.Email, &u.Name, &u.PasswordHash,
		&u.Status, &u.EmailVerified, &u.AvatarURL,
		&metadataPublicBytes, &metadataPrivateBytes,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt, &u.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("not found: %w", err)
		}
		return nil, err
	}

	u.MetadataPublic = make(map[string]interface{})
	if len(metadataPublicBytes) > 0 {
		if err := json.Unmarshal(metadataPublicBytes, &u.MetadataPublic); err != nil {
			return nil, err
		}
	}

	u.MetadataPrivate = make(map[string]interface{})
	if len(metadataPrivateBytes) > 0 {
		if err := json.Unmarshal(metadataPrivateBytes, &u.MetadataPrivate); err != nil {
			return nil, err
		}
	}

	return &u, nil
}
