package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/medaminerjb/saas-kit/internal/identity/domain"
	"github.com/medaminerjb/saas-kit/internal/identity/repository"
	"github.com/medaminerjb/saas-kit/internal/platform/events"
)

// APIKeyService handles API key operations.
type APIKeyService struct {
	repo     repository.APIKeyRepository
	publisher events.Publisher
	logger   *slog.Logger
}

// NewAPIKeyService creates a new APIKeyService.
func NewAPIKeyService(repo repository.APIKeyRepository, publisher events.Publisher, logger *slog.Logger) *APIKeyService {
	return &APIKeyService{
		repo:     repo,
		publisher: publisher,
		logger:   logger,
	}
}

// CreateKey creates a new API key for a tenant.
func (s *APIKeyService) CreateKey(ctx context.Context, tenantID uuid.UUID, name string, keyType domain.APIKeyType, scopes []string, expiresAt *time.Time, createdBy uuid.UUID) (*domain.APIKey, string, error) {
	if name == "" {
		return nil, "", fmt.Errorf("name is required")
	}

	if len(scopes) == 0 {
		return nil, "", fmt.Errorf("at least one scope is required")
	}

	// Generate the full API key (32 random bytes = 64 hex chars)
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	fullKey := hex.EncodeToString(randomBytes)

	// Generate prefix based on type
	prefix := "sk_test_"
	if keyType == domain.APIKeyTypeLive {
		prefix = "sk_live_"
	}

	// Hash the full key for storage
	hash := sha256.Sum256([]byte(fullKey))
	keyHash := hex.EncodeToString(hash[:])

	// Create the API key domain object
	key := &domain.APIKey{
		TenantID:  tenantID,
		Name:      name,
		KeyPrefix: prefix,
		KeyHash:   keyHash,
		Scopes:    scopes,
		Type:      keyType,
		Status:    domain.APIKeyStatusActive,
		ExpiresAt: expiresAt,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, key); err != nil {
		return nil, "", fmt.Errorf("failed to create API key: %w", err)
	}

	// The full key to return to the user (only shown once)
	displayKey := prefix + fullKey

	s.logger.InfoContext(ctx, "API key created", "tenant_id", tenantID, "name", name, "type", keyType)

	return key, displayKey, nil
}

// GetKey retrieves an API key by ID.
func (s *APIKeyService) GetKey(ctx context.Context, id, tenantID uuid.UUID) (*domain.APIKey, error) {
	return s.repo.GetByID(ctx, id, tenantID)
}

// ListKeys retrieves all API keys for a tenant.
func (s *APIKeyService) ListKeys(ctx context.Context, tenantID uuid.UUID) ([]*domain.APIKey, error) {
	return s.repo.ListByTenant(ctx, tenantID)
}

// RevokeKey revokes an API key.
func (s *APIKeyService) RevokeKey(ctx context.Context, id, tenantID, revokedBy uuid.UUID) error {
	if err := s.repo.Revoke(ctx, id, tenantID, revokedBy); err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	s.logger.InfoContext(ctx, "API key revoked", "id", id, "tenant_id", tenantID, "revoked_by", revokedBy)

	return nil
}

// DeleteKey permanently deletes an API key.
func (s *APIKeyService) DeleteKey(ctx context.Context, id, tenantID uuid.UUID) error {
	if err := s.repo.Delete(ctx, id, tenantID); err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	s.logger.InfoContext(ctx, "API key deleted", "id", id, "tenant_id", tenantID)

	return nil
}

// ValidateKey validates an API key and returns the key if valid.
func (s *APIKeyService) ValidateKey(ctx context.Context, fullKey string) (*domain.APIKey, error) {
	// Extract the prefix and hash the full key
	hash := sha256.Sum256([]byte(fullKey))
	keyHash := hex.EncodeToString(hash[:])

	key, err := s.repo.GetByHash(ctx, keyHash)
	if err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	// Check if the key is active and not expired
	if !key.IsActive() {
		return nil, fmt.Errorf("API key is not active or has expired")
	}

	// Update last used timestamp
	if err := s.repo.UpdateLastUsed(ctx, key.ID); err != nil {
		s.logger.WarnContext(ctx, "failed to update API key last used", "id", key.ID, "error", err)
	}

	return key, nil
}
