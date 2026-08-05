// Package domain defines the core identity domain models.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// APIKeyStatus represents the lifecycle state of an API key.
type APIKeyStatus string

const (
	APIKeyStatusActive   APIKeyStatus = "active"
	APIKeyStatusRevoked  APIKeyStatus = "revoked"
	APIKeyStatusExpired  APIKeyStatus = "expired"
)

// APIKeyType represents the type of API key (test vs live).
type APIKeyType string

const (
	APIKeyTypeTest APIKeyType = "test"
	APIKeyTypeLive APIKeyType = "live"
)

// APIKey represents a machine-to-machine authentication credential.
type APIKey struct {
	ID           uuid.UUID      `json:"id"`
	TenantID     uuid.UUID      `json:"tenant_id"`
	Name         string         `json:"name"`
	KeyPrefix    string         `json:"key_prefix"` // sk_live_ or sk_test_
	KeyHash      string         `json:"-"`         // SHA-256 hash of the full key
	Scopes       []string       `json:"scopes"`
	Type         APIKeyType     `json:"type"`
	Status       APIKeyStatus   `json:"status"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
	LastUsedAt   *time.Time     `json:"last_used_at,omitempty"`
	CreatedBy    uuid.UUID      `json:"created_by"`
	CreatedAt    time.Time      `json:"created_at"`
	RevokedAt    *time.Time     `json:"revoked_at,omitempty"`
	RevokedBy    *uuid.UUID     `json:"revoked_by,omitempty"`
}

// IsActive returns true if the API key is active and not expired.
func (k *APIKey) IsActive() bool {
	if k.Status != APIKeyStatusActive {
		return false
	}
	if k.ExpiresAt != nil && k.ExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}

// HasScope returns true if the API key has the specified scope.
func (k *APIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}
