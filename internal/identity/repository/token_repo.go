package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/saaskit/saaskit/internal/identity/domain"
)

// TokenRepo implements repository.IdentityTokenRepository using pgx.
type TokenRepo struct {
	pool *pgxpool.Pool
}

// NewTokenRepo creates a new PostgreSQL identity token repository.
func NewTokenRepo(pool *pgxpool.Pool) *TokenRepo {
	return &TokenRepo{pool: pool}
}

// Create inserts a new identity token.
func (r *TokenRepo) Create(ctx context.Context, token *domain.IdentityToken) error {
	query := `INSERT INTO identity_tokens (id, user_id, type, hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, query,
		token.ID, token.UserID, token.Type, token.Hash,
		token.ExpiresAt, token.CreatedAt,
	)
	return err
}

// GetByHash retrieves an unused, non-expired token by its HMAC hash.
func (r *TokenRepo) GetByHash(ctx context.Context, hash string) (*domain.IdentityToken, error) {
	query := `SELECT id, user_id, type, hash, expires_at, used_at, created_at
		FROM identity_tokens WHERE hash = $1 AND used_at IS NULL AND expires_at > NOW()`

	var t domain.IdentityToken
	err := r.pool.QueryRow(ctx, query, hash).Scan(
		&t.ID, &t.UserID, &t.Type, &t.Hash,
		&t.ExpiresAt, &t.UsedAt, &t.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("token not found: %w", err)
		}
		return nil, err
	}
	return &t, nil
}

// MarkUsed marks a token as used.
func (r *TokenRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE identity_tokens SET used_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// InvalidateByUserAndType marks all unused tokens of a given type for a user as used.
func (r *TokenRepo) InvalidateByUserAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) error {
	query := `UPDATE identity_tokens SET used_at = NOW() WHERE user_id = $1 AND type = $2 AND used_at IS NULL`
	_, err := r.pool.Exec(ctx, query, userID, tokenType)
	return err
}

// DeleteExpired removes expired and used tokens.
func (r *TokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM identity_tokens WHERE expires_at < NOW() OR used_at IS NOT NULL`
	ct, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}
