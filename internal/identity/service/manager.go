// Package service provides the IdentityManager, an orchestration layer
// that coordinates identity sub-services (auth, users, tokens, sessions).
//
// As the platform grows to include MFA, WebAuthn, passkeys, device trust,
// risk detection, and session policies, the IdentityManager will orchestrate
// those flows without bloating individual services.
package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/medaminerjb/saas-kit/internal/identity/domain"
)

// IdentityManager is the top-level orchestrator for identity operations.
// It delegates to specialized services and coordinates cross-cutting flows.
type IdentityManager struct {
	Auth   *AuthService
	Users  *UserService
	Tokens *TokenService
	logger *slog.Logger
}

// NewIdentityManager creates a new identity orchestrator.
func NewIdentityManager(auth *AuthService, users *UserService, tokens *TokenService, logger *slog.Logger) *IdentityManager {
	return &IdentityManager{
		Auth:   auth,
		Users:  users,
		Tokens: tokens,
		logger: logger,
	}
}

// Register orchestrates user registration.
// In the future this will also trigger: MFA enrollment prompts,
// welcome emails, default tenant creation, etc.
func (m *IdentityManager) Register(ctx context.Context, input RegisterInput) (*domain.User, *AuthTokens, error) {
	return m.Auth.Register(ctx, input)
}

// Login orchestrates user login.
// Future: MFA challenge, device trust check, risk assessment.
func (m *IdentityManager) Login(ctx context.Context, input LoginInput) (*domain.User, *AuthTokens, error) {
	return m.Auth.Login(ctx, input)
}

// RefreshTokens orchestrates token refresh.
func (m *IdentityManager) RefreshTokens(ctx context.Context, refreshToken string) (*AuthTokens, error) {
	return m.Auth.RefreshTokens(ctx, refreshToken)
}

// Logout orchestrates session termination.
func (m *IdentityManager) Logout(ctx context.Context, sessionID, userID uuid.UUID) error {
	return m.Auth.Logout(ctx, sessionID, userID)
}

// GetCurrentUser retrieves the authenticated user's profile.
func (m *IdentityManager) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return m.Users.GetByID(ctx, userID)
}

// UpdateProfile updates the authenticated user's profile.
func (m *IdentityManager) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*domain.User, error) {
	return m.Users.UpdateProfile(ctx, userID, input)
}
