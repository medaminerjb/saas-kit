package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medaminerjb/saas-kit/internal/identity/domain"
	"github.com/medaminerjb/saas-kit/internal/sqlcgen"
)

// APIKeyRepository defines the interface for API key persistence.
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.APIKey, error)
	GetByHash(ctx context.Context, hash string) (*domain.APIKey, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.APIKey, error)
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id, tenantID, revokedBy uuid.UUID) error
	Delete(ctx context.Context, id, tenantID uuid.UUID) error
}

// APIKeyRepo implements APIKeyRepository using sqlc-generated queries.
type APIKeyRepo struct {
	pool    *pgxpool.Pool
	queries *sqlcgen.Queries
}

// NewAPIKeyRepo creates a new APIKeyRepo.
func NewAPIKeyRepo(pool *pgxpool.Pool) *APIKeyRepo {
	return &APIKeyRepo{
		pool:    pool,
		queries: sqlcgen.New(pool),
	}
}

func (r *APIKeyRepo) Create(ctx context.Context, key *domain.APIKey) error {
	var expiresAt pgtype.Timestamptz
	if key.ExpiresAt != nil {
		expiresAt = pgtype.Timestamptz{Time: *key.ExpiresAt, Valid: true}
	}

	params := sqlcgen.CreateAPIKeyParams{
		TenantID:  key.TenantID,
		Name:      key.Name,
		KeyPrefix: key.KeyPrefix,
		KeyHash:   key.KeyHash,
		Scopes:    key.Scopes,
		Type:      string(key.Type),
		Status:    string(key.Status),
		ExpiresAt: expiresAt,
		CreatedBy: key.CreatedBy,
	}

	result, err := r.queries.CreateAPIKey(ctx, params)
	if err != nil {
		return err
	}

	*key = r.sqlcToDomain(result)
	return nil
}

func (r *APIKeyRepo) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.APIKey, error) {
	result, err := r.queries.GetAPIKeyByID(ctx, sqlcgen.GetAPIKeyByIDParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}

	key := r.sqlcToDomain(result)
	return &key, nil
}

func (r *APIKeyRepo) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	result, err := r.queries.GetAPIKeyByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	key := r.sqlcToDomain(result)
	return &key, nil
}

func (r *APIKeyRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.APIKey, error) {
	results, err := r.queries.ListAPIKeysByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	keys := make([]*domain.APIKey, len(results))
	for i, result := range results {
		key := r.sqlcToDomain(result)
		keys[i] = &key
	}

	return keys, nil
}

func (r *APIKeyRepo) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.queries.UpdateAPIKeyLastUsed(ctx, id)
	return err
}

func (r *APIKeyRepo) Revoke(ctx context.Context, id, tenantID, revokedBy uuid.UUID) error {
	revokedByPG := pgtype.UUID{Bytes: revokedBy, Valid: true}
	_, err := r.queries.RevokeAPIKey(ctx, sqlcgen.RevokeAPIKeyParams{
		ID:        id,
		RevokedBy: revokedByPG,
	})
	return err
}

func (r *APIKeyRepo) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	return r.queries.DeleteAPIKey(ctx, sqlcgen.DeleteAPIKeyParams{
		ID:       id,
		TenantID: tenantID,
	})
}

func (r *APIKeyRepo) sqlcToDomain(result sqlcgen.ApiKey) domain.APIKey {
	var expiresAt, lastUsedAt, revokedAt *time.Time
	if result.ExpiresAt.Valid {
		expiresAt = &result.ExpiresAt.Time
	}
	if result.LastUsedAt.Valid {
		lastUsedAt = &result.LastUsedAt.Time
	}
	if result.RevokedAt.Valid {
		revokedAt = &result.RevokedAt.Time
	}

	var revokedBy *uuid.UUID
	if result.RevokedBy.Valid {
		revokedByUUID := uuid.UUID(result.RevokedBy.Bytes)
		revokedBy = &revokedByUUID
	}

	return domain.APIKey{
		ID:         result.ID,
		TenantID:   result.TenantID,
		Name:       result.Name,
		KeyPrefix:  result.KeyPrefix,
		KeyHash:    result.KeyHash,
		Scopes:     result.Scopes,
		Type:       domain.APIKeyType(result.Type),
		Status:     domain.APIKeyStatus(result.Status),
		ExpiresAt:  expiresAt,
		LastUsedAt: lastUsedAt,
		CreatedBy:  result.CreatedBy,
		CreatedAt:  result.CreatedAt,
		RevokedAt:  revokedAt,
		RevokedBy:  revokedBy,
	}
}
