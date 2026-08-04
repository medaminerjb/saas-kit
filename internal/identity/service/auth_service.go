package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	idcrypto "github.com/medaminerjb/saas-kit/internal/identity/crypto"
	"github.com/medaminerjb/saas-kit/internal/identity/domain"
	"github.com/medaminerjb/saas-kit/internal/identity/repository"
	"github.com/medaminerjb/saas-kit/internal/platform/events"
)

// AuthService handles authentication flows: registration, login, token refresh, logout.
type AuthService struct {
	users        repository.UserRepository
	sessions     repository.SessionRepository
	tokens       repository.IdentityTokenRepository
	hasher       *idcrypto.Hasher
	tokenHasher  *idcrypto.TokenHasher
	tokenService *TokenService
	publisher    events.Publisher
	refreshTTL   time.Duration
	logger       *slog.Logger

	failedMu       sync.Mutex
	failedAttempts map[string]int
}

// AuthServiceConfig holds AuthService dependencies.
type AuthServiceConfig struct {
	Users        repository.UserRepository
	Sessions     repository.SessionRepository
	Tokens       repository.IdentityTokenRepository
	Hasher       *idcrypto.Hasher
	TokenHasher  *idcrypto.TokenHasher
	TokenService *TokenService
	Publisher    events.Publisher
	RefreshTTL   time.Duration
}

// NewAuthService creates a new authentication service.
func NewAuthService(cfg AuthServiceConfig, logger *slog.Logger) *AuthService {
	return &AuthService{
		users:          cfg.Users,
		sessions:       cfg.Sessions,
		tokens:         cfg.Tokens,
		hasher:         cfg.Hasher,
		tokenHasher:    cfg.TokenHasher,
		tokenService:   cfg.TokenService,
		publisher:      cfg.Publisher,
		refreshTTL:     cfg.RefreshTTL,
		logger:         logger,
		failedAttempts: make(map[string]int),
	}
}

// RegisterInput holds the data needed to register a new user.
type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

// AuthTokens holds the access and refresh tokens returned after authentication.
type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// Register creates a new user account with email/password.
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*domain.User, *AuthTokens, error) {
	if err := ValidateEmail(input.Email); err != nil {
		return nil, nil, err
	}
	if err := ValidatePassword(input.Password); err != nil {
		return nil, nil, err
	}

	// Check for existing user
	existing, _ := s.users.GetByEmail(ctx, input.Email, nil)
	if existing != nil {
		return nil, nil, domain.ErrUserAlreadyExists
	}

	// Hash password
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &domain.User{
		ID:            uuid.New(),
		Email:         input.Email,
		Name:          input.Name,
		PasswordHash:  &passwordHash,
		Status:        domain.UserStatusPendingVerification,
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("creating user: %w", err)
	}

	// Create session and tokens
	tokens, err := s.createSession(ctx, user, "", "")
	if err != nil {
		return nil, nil, err
	}

	// Publish event
	_ = s.publisher.Publish(ctx, events.Event{
		Type:      "user.registered",
		ActorID:   &user.ID,
		TargetID:  &user.ID,
		Timestamp: time.Now(),
	})

	s.logger.InfoContext(ctx, "user registered",
		slog.String("user_id", user.ID.String()),
		slog.String("email", user.Email),
	)

	return user, tokens, nil
}

// LoginInput holds the data needed for email/password login.
type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

// Login authenticates a user with email/password.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*domain.User, *AuthTokens, error) {
	user, err := s.users.GetByEmail(ctx, input.Email, nil)
	if err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	if !user.CanLogin() {
		switch user.Status {
		case domain.UserStatusDisabled:
			return nil, nil, domain.ErrAccountDisabled
		case domain.UserStatusLocked:
			return nil, nil, domain.ErrAccountLocked
		default:
			return nil, nil, domain.ErrInvalidCredentials
		}
	}

	if !user.HasPassword() {
		return nil, nil, domain.ErrInvalidCredentials
	}

	match, err := s.hasher.Verify(input.Password, *user.PasswordHash)
	if err != nil || !match {
		s.trackFailedLogin(ctx, user.Email, user)
		return nil, nil, domain.ErrInvalidCredentials
	}

	// Reset attempts on success
	s.resetFailedLogin(user.Email)

	// Update last login
	_ = s.users.UpdateLastLogin(ctx, user.ID)

	// Create session
	tokens, err := s.createSession(ctx, user, input.UserAgent, input.IPAddress)
	if err != nil {
		return nil, nil, err
	}

	_ = s.publisher.Publish(ctx, events.Event{
		Type:      "user.login",
		ActorID:   &user.ID,
		TargetID:  &user.ID,
		Timestamp: time.Now(),
	})

	return user, tokens, nil
}

// RefreshTokens validates a refresh token and issues a new token pair (rotation).
// Implements 10-second grace window for concurrent refresh requests.
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*AuthTokens, error) {
	hash := s.tokenHasher.Hash(refreshToken)

	session, err := s.sessions.GetByRefreshTokenHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	if !session.IsValid() {
		return nil, domain.ErrSessionExpired
	}

	// Store previous token hash and rotation timestamp for grace window
	now := time.Now()
	previousHash := session.RefreshTokenHash
	session.PreviousRefreshTokenHash = &previousHash
	session.RotatedAt = &now

	// Update session with grace window data before creating new one
	// Note: We don't revoke immediately - the grace window allows the old token
	// to work for 10 seconds after rotation
	if err := s.sessions.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("updating session for grace window: %w", err)
	}

	// Load user
	user, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("loading user for refresh: %w", err)
	}

	if !user.CanLogin() {
		return nil, domain.ErrAccountDisabled
	}

	// Create new session with new refresh token
	tokens, err := s.createSession(ctx, user, session.UserAgent, session.IPAddress)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// Logout revokes a session by its ID.
func (s *AuthService) Logout(ctx context.Context, sessionID, userID uuid.UUID) error {
	if err := s.sessions.Revoke(ctx, sessionID); err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}

	_ = s.publisher.Publish(ctx, events.Event{
		Type:      "user.logout",
		ActorID:   &userID,
		Timestamp: time.Now(),
	})

	return nil
}

// RequestPasswordReset generates a password reset token.
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.users.GetByEmail(ctx, email, nil)
	if err != nil {
		// Don't reveal whether the user exists
		return nil
	}

	// Invalidate any existing reset tokens for this user
	_ = s.tokens.InvalidateByUserAndType(ctx, user.ID, domain.TokenTypePasswordReset)

	// Generate new token
	raw, hash, err := s.tokenHasher.GenerateToken(32)
	if err != nil {
		return fmt.Errorf("generating reset token: %w", err)
	}

	token := &domain.IdentityToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      domain.TokenTypePasswordReset,
		Hash:      hash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.tokens.Create(ctx, token); err != nil {
		return fmt.Errorf("saving reset token: %w", err)
	}

	_ = s.publisher.Publish(ctx, events.Event{
		Type:      "user.password_reset_requested",
		ActorID:   &user.ID,
		TargetID:  &user.ID,
		Payload:   map[string]string{"token": raw}, // Consumed by email sender
		Timestamp: time.Now(),
	})

	return nil
}

// ResetPassword validates a reset token and sets a new password.
func (s *AuthService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	hash := s.tokenHasher.Hash(rawToken)
	token, err := s.tokens.GetByHash(ctx, hash)
	if err != nil || token == nil {
		return domain.ErrInvalidToken
	}

	if !token.IsValid() || token.Type != domain.TokenTypePasswordReset {
		return domain.ErrInvalidToken
	}

	// Hash new password
	passwordHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	// Update password
	if err := s.users.UpdatePassword(ctx, token.UserID, passwordHash); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	// Mark token as used
	if err := s.tokens.MarkUsed(ctx, token.ID); err != nil {
		return fmt.Errorf("marking token used: %w", err)
	}

	// Revoke all sessions (force re-login)
	_ = s.sessions.RevokeAllForUser(ctx, token.UserID)

	_ = s.publisher.Publish(ctx, events.Event{
		Type:      "user.password_changed",
		ActorID:   &token.UserID,
		TargetID:  &token.UserID,
		Timestamp: time.Now(),
	})

	return nil
}

// VerifyEmail validates an email verification token.
func (s *AuthService) VerifyEmail(ctx context.Context, rawToken string) error {
	hash := s.tokenHasher.Hash(rawToken)
	token, err := s.tokens.GetByHash(ctx, hash)
	if err != nil || token == nil {
		return domain.ErrInvalidToken
	}

	if !token.IsValid() || token.Type != domain.TokenTypeEmailVerification {
		return domain.ErrInvalidToken
	}

	if err := s.users.SetEmailVerified(ctx, token.UserID); err != nil {
		return fmt.Errorf("setting email verified: %w", err)
	}

	if err := s.tokens.MarkUsed(ctx, token.ID); err != nil {
		return fmt.Errorf("marking token used: %w", err)
	}

	_ = s.publisher.Publish(ctx, events.Event{
		Type:      "user.email_verified",
		ActorID:   &token.UserID,
		TargetID:  &token.UserID,
		Timestamp: time.Now(),
	})

	return nil
}

// createSession generates a refresh token, stores the session, and returns a token pair.
func (s *AuthService) createSession(ctx context.Context, user *domain.User, userAgent, ipAddress string) (*AuthTokens, error) {
	// Generate refresh token
	rawRefresh, refreshHash, err := s.tokenHasher.GenerateToken(48)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	session := &domain.Session{
		ID:               uuid.New(),
		UserID:           user.ID,
		TenantID:         user.TenantID,
		RefreshTokenHash: refreshHash,
		UserAgent:        userAgent,
		IPAddress:        ipAddress,
		ExpiresAt:        time.Now().Add(s.refreshTTL),
		CreatedAt:        time.Now(),
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	// Generate access token
	accessToken, err := s.tokenService.GenerateAccessToken(ctx, user, session.ID, nil, nil, false)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.tokenService.accessTTL.Seconds()),
	}, nil
}

// CreateSessionForUser creates a session and tokens for a pre-authenticated user
// (e.g., after social login where credentials were verified by the external provider).
func (s *AuthService) CreateSessionForUser(ctx context.Context, user *domain.User) (*AuthTokens, error) {
	return s.createSession(ctx, user, "social-login", "")
}

func (s *AuthService) trackFailedLogin(ctx context.Context, email string, user *domain.User) {
	s.failedMu.Lock()
	defer s.failedMu.Unlock()

	s.failedAttempts[email]++
	attempts := s.failedAttempts[email]

	s.logger.WarnContext(ctx, "failed login attempt",
		slog.String("email", email),
		slog.Int("attempts", attempts),
	)

	if attempts >= 5 {
		user.Status = domain.UserStatusLocked
		if err := s.users.Update(ctx, user); err != nil {
			s.logger.ErrorContext(ctx, "failed to lock user account", slog.Any("error", err))
		} else {
			s.logger.WarnContext(ctx, "user account locked due to excessive failed login attempts",
				slog.String("email", email),
				slog.String("user_id", user.ID.String()),
			)
			_ = s.publisher.Publish(ctx, events.Event{
				Type:     "user.locked",
				ActorID:  &user.ID,
				TargetID: &user.ID,
			})
		}
		delete(s.failedAttempts, email)
	}
}

func (s *AuthService) resetFailedLogin(email string) {
	s.failedMu.Lock()
	defer s.failedMu.Unlock()
	delete(s.failedAttempts, email)
}
