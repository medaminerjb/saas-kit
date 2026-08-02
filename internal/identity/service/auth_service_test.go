package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/medaminerjb/saas-kit/internal/identity/crypto"
	"github.com/medaminerjb/saas-kit/internal/identity/domain"
)

// mockSessionRepository implements repository.SessionRepository for testing.
type mockSessionRepository struct {
	createFunc                func(ctx context.Context, session *domain.Session) error
	getByRefreshTokenHashFunc func(ctx context.Context, hash string) (*domain.Session, error)
	revokeFunc                func(ctx context.Context, id uuid.UUID) error
}

func (m *mockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, session)
	}
	return nil
}

func (m *mockSessionRepository) GetByRefreshTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	if m.getByRefreshTokenHashFunc != nil {
		return m.getByRefreshTokenHashFunc(ctx, hash)
	}
	return nil, nil
}

func (m *mockSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	return nil, nil
}

func (m *mockSessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	if m.revokeFunc != nil {
		return m.revokeFunc(ctx, id)
	}
	return nil
}

func (m *mockSessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *mockSessionRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	return nil, nil
}

func (m *mockSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

// mockIdentityTokenRepository implements repository.IdentityTokenRepository for testing.
type mockIdentityTokenRepository struct{}

func (m *mockIdentityTokenRepository) Create(ctx context.Context, token *domain.IdentityToken) error {
	return nil
}

func (m *mockIdentityTokenRepository) GetByHash(ctx context.Context, hash string) (*domain.IdentityToken, error) {
	return nil, nil
}

func (m *mockIdentityTokenRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockIdentityTokenRepository) InvalidateByUserAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) error {
	return nil
}

func (m *mockIdentityTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

func TestAuthService_LoginAndLockout(t *testing.T) {
	tmpDir := t.TempDir()
	kp, err := crypto.LoadOrGenerateKeyPair(tmpDir, "RS256", true)
	if err != nil {
		t.Fatalf("key generation: %v", err)
	}

	tokenService := NewTokenService(TokenServiceConfig{
		PrivateKey: kp.PrivateKey,
		PublicKey:  kp.PublicKey,
		Algorithm:  "RS256",
		KeyID:      "test-key",
		Issuer:     "https://test.saaskit.dev",
		AccessTTL:  15 * time.Minute,
	}, nil)

	hasher := crypto.NewHasher(crypto.Argon2Params{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	})

	tokenHasher := crypto.NewTokenHasher("test-server-secret")

	password := "Secret123!"
	pwHash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	uID := uuid.New()
	user := &domain.User{
		ID:            uID,
		Email:         "test@example.com",
		PasswordHash:  &pwHash,
		Status:        domain.UserStatusActive,
		EmailVerified: true,
	}

	mockRepo := &mockUserRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
		getByEmailFunc: func(ctx context.Context, email string, tenantID *uuid.UUID) (*domain.User, error) {
			if email == user.Email {
				return user, nil
			}
			return nil, errors.New("not found")
		},
		updateFunc: func(ctx context.Context, u *domain.User) error {
			user = u
			return nil
		},
	}

	mockSessions := &mockSessionRepository{}
	mockTokens := &mockIdentityTokenRepository{}
	pub := &mockPublisher{}

	authSvc := NewAuthService(AuthServiceConfig{
		Users:        mockRepo,
		Sessions:     mockSessions,
		Tokens:       mockTokens,
		Hasher:       hasher,
		TokenHasher:  tokenHasher,
		TokenService: tokenService,
		Publisher:    pub,
		RefreshTTL:   7 * 24 * time.Hour,
	}, slog.Default())

	t.Run("successful login", func(t *testing.T) {
		resUser, tokens, err := authSvc.Login(context.Background(), LoginInput{
			Email:    "test@example.com",
			Password: password,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resUser.Email != "test@example.com" {
			t.Errorf("expected user email to be test@example.com, got %s", resUser.Email)
		}
		if tokens.AccessToken == "" {
			t.Error("expected access token to be generated")
		}
	})

	t.Run("failed login lockout flow", func(t *testing.T) {
		// Reset user status to active
		user.Status = domain.UserStatusActive

		// Make 4 failed attempts
		for i := 0; i < 4; i++ {
			_, _, err := authSvc.Login(context.Background(), LoginInput{
				Email:    "test@example.com",
				Password: "WrongPassword",
			})
			if !errors.Is(err, domain.ErrInvalidCredentials) {
				t.Errorf("attempt %d: expected ErrInvalidCredentials, got %v", i+1, err)
			}
			if user.Status != domain.UserStatusActive {
				t.Errorf("attempt %d: expected user status to remain active, got %s", i+1, user.Status)
			}
		}

		// 5th failed attempt should lock the user account
		_, _, err := authSvc.Login(context.Background(), LoginInput{
			Email:    "test@example.com",
			Password: "WrongPassword",
		})
		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Errorf("5th attempt: expected ErrInvalidCredentials, got %v", err)
		}
		if user.Status != domain.UserStatusLocked {
			t.Errorf("expected user status to be locked, got %s", user.Status)
		}

		// 6th attempt (even with correct password) should return ErrAccountLocked
		_, _, err = authSvc.Login(context.Background(), LoginInput{
			Email:    "test@example.com",
			Password: password,
		})
		if !errors.Is(err, domain.ErrAccountLocked) {
			t.Errorf("expected ErrAccountLocked, got %v", err)
		}
	})
}

func TestAuthService_Logout(t *testing.T) {
	mockSessions := &mockSessionRepository{}
	var revokedID uuid.UUID
	mockSessions.revokeFunc = func(ctx context.Context, id uuid.UUID) error {
		revokedID = id
		return nil
	}

	authSvc := NewAuthService(AuthServiceConfig{
		Sessions:  mockSessions,
		Publisher: &mockPublisher{},
	}, slog.Default())

	sessionID := uuid.New()
	userID := uuid.New()
	err := authSvc.Logout(context.Background(), sessionID, userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if revokedID != sessionID {
		t.Errorf("expected session %s to be revoked, got %s", sessionID, revokedID)
	}
}

func TestAuthService_RefreshTokens(t *testing.T) {
	tmpDir := t.TempDir()
	kp, err := crypto.LoadOrGenerateKeyPair(tmpDir, "RS256", true)
	if err != nil {
		t.Fatalf("key generation: %v", err)
	}

	tokenService := NewTokenService(TokenServiceConfig{
		PrivateKey: kp.PrivateKey,
		PublicKey:  kp.PublicKey,
		Algorithm:  "RS256",
		KeyID:      "test-key",
		Issuer:     "https://test.saaskit.dev",
		AccessTTL:  15 * time.Minute,
	}, nil)

	hasher := crypto.NewHasher(crypto.Argon2Params{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	})

	tokenHasher := crypto.NewTokenHasher("test-server-secret")

	uID := uuid.New()
	user := &domain.User{
		ID:            uID,
		Email:         "test@example.com",
		Status:        domain.UserStatusActive,
		EmailVerified: true,
	}

	mockRepo := &mockUserRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			if id == uID {
				return user, nil
			}
			return nil, errors.New("not found")
		},
	}

	mockSessions := &mockSessionRepository{}
	session := &domain.Session{
		ID:               uuid.New(),
		UserID:           uID,
		RefreshTokenHash: tokenHasher.Hash("valid-refresh-token"),
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	}

	mockSessions.getByRefreshTokenHashFunc = func(ctx context.Context, hash string) (*domain.Session, error) {
		if hash == session.RefreshTokenHash {
			return session, nil
		}
		return nil, errors.New("not found")
	}

	var oldSessionRevoked bool
	mockSessions.revokeFunc = func(ctx context.Context, id uuid.UUID) error {
		if id == session.ID {
			oldSessionRevoked = true
		}
		return nil
	}

	mockTokens := &mockIdentityTokenRepository{}
	pub := &mockPublisher{}

	authSvc := NewAuthService(AuthServiceConfig{
		Users:        mockRepo,
		Sessions:     mockSessions,
		Tokens:       mockTokens,
		Hasher:       hasher,
		TokenHasher:  tokenHasher,
		TokenService: tokenService,
		Publisher:    pub,
		RefreshTTL:   7 * 24 * time.Hour,
	}, slog.Default())

	tokens, err := authSvc.RefreshTokens(context.Background(), "valid-refresh-token")
	if err != nil {
		t.Fatalf("expected no error on refresh, got %v", err)
	}

	if !oldSessionRevoked {
		t.Error("expected old session to be revoked during refresh token rotation")
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Error("expected new tokens to be generated")
	}
}

