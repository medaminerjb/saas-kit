// Package repository defines the persistence interfaces for the identity module.
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/medaminerjb/saas-kit/internal/identity/domain"
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string, tenantID *uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateMetadata(ctx context.Context, id uuid.UUID, metadataPublic, metadataPrivate map[string]interface{}) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	SetEmailVerified(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, tenantID *uuid.UUID, limit, offset int32) ([]*domain.User, error)
	Count(ctx context.Context, tenantID *uuid.UUID) (int64, error)
}

// SessionRepository defines persistence operations for sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByRefreshTokenHash(ctx context.Context, hash string) (*domain.Session, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	Update(ctx context.Context, session *domain.Session) error
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	ListForUser(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error)
	DeleteExpired(ctx context.Context) (int64, error)
}

// IdentityTokenRepository defines persistence operations for identity tokens.
type IdentityTokenRepository interface {
	Create(ctx context.Context, token *domain.IdentityToken) error
	GetByHash(ctx context.Context, hash string) (*domain.IdentityToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	InvalidateByUserAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) error
	DeleteExpired(ctx context.Context) (int64, error)
}
