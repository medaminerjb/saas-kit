package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/saaskit/saaskit/internal/identity/domain"
)

// SessionRepo implements repository.SessionRepository using pgx.
type SessionRepo struct {
	pool *pgxpool.Pool
}

// NewSessionRepo creates a new PostgreSQL session repository.
func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{pool: pool}
}

// Create inserts a new session.
func (r *SessionRepo) Create(ctx context.Context, session *domain.Session) error {
	query := `INSERT INTO sessions (id, user_id, tenant_id, refresh_token_hash, user_agent, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6::inet, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		session.ID, session.UserID, session.TenantID,
		session.RefreshTokenHash, session.UserAgent,
		nullableIP(session.IPAddress), session.ExpiresAt, session.CreatedAt,
	)
	return err
}

// GetByRefreshTokenHash retrieves an active session by its refresh token hash.
func (r *SessionRepo) GetByRefreshTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	query := `SELECT id, user_id, tenant_id, refresh_token_hash, user_agent, ip_address, expires_at, created_at, revoked_at
		FROM sessions WHERE refresh_token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()`

	return r.scanSession(r.pool.QueryRow(ctx, query, hash))
}

// GetByID retrieves a session by ID.
func (r *SessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	query := `SELECT id, user_id, tenant_id, refresh_token_hash, user_agent, ip_address, expires_at, created_at, revoked_at
		FROM sessions WHERE id = $1`

	return r.scanSession(r.pool.QueryRow(ctx, query, id))
}

// Revoke marks a session as revoked.
func (r *SessionRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sessions SET revoked_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// RevokeAllForUser revokes all active sessions for a user.
func (r *SessionRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE sessions SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// ListForUser returns all active sessions for a user.
func (r *SessionRepo) ListForUser(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	query := `SELECT id, user_id, tenant_id, refresh_token_hash, user_agent, ip_address, expires_at, created_at, revoked_at
		FROM sessions WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		s, err := r.scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// DeleteExpired removes expired and revoked sessions.
func (r *SessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < NOW() OR revoked_at IS NOT NULL`
	ct, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

func (r *SessionRepo) scanSession(row scannable) (*domain.Session, error) {
	var s domain.Session
	var ipAddress *string
	err := row.Scan(
		&s.ID, &s.UserID, &s.TenantID, &s.RefreshTokenHash,
		&s.UserAgent, &ipAddress, &s.ExpiresAt, &s.CreatedAt, &s.RevokedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("session not found: %w", err)
		}
		return nil, err
	}
	if ipAddress != nil {
		s.IPAddress = *ipAddress
	}
	return &s, nil
}

// nullableIP returns nil for empty IP strings (to avoid INET parse errors).
func nullableIP(ip string) any {
	if ip == "" {
		return nil
	}
	return ip
}
