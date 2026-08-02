package service

import (
	"context"
	"testing"
	"time"

	"github.com/medaminerjb/saas-kit/internal/identity/crypto"
	"github.com/medaminerjb/saas-kit/internal/identity/domain"
)

func TestTokenService_GenerateAndValidate_RS256(t *testing.T) {
	tmpDir := t.TempDir()
	kp, err := crypto.LoadOrGenerateKeyPair(tmpDir, "RS256", true)
	if err != nil {
		t.Fatalf("key generation: %v", err)
	}

	svc := NewTokenService(TokenServiceConfig{
		PrivateKey: kp.PrivateKey,
		PublicKey:  kp.PublicKey,
		Algorithm:  "RS256",
		KeyID:      "test-key",
		Issuer:     "https://test.saaskit.dev",
		AccessTTL:  15 * time.Minute,
	}, nil)

	user := &domain.User{
		Email:  "test@example.com",
		Name:   "Test User",
		Status: domain.UserStatusActive,
	}

	ctx := context.Background()
	token, err := svc.GenerateAccessToken(ctx, user, [16]byte{})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error: %v", err)
	}

	if token == "" {
		t.Error("token should not be empty")
	}

	// Validate the token
	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error: %v", err)
	}

	if claims.Issuer != "https://test.saaskit.dev" {
		t.Errorf("issuer = %q, want https://test.saaskit.dev", claims.Issuer)
	}

	if claims.Scope != "openid profile email" {
		t.Errorf("scope = %q, want 'openid profile email'", claims.Scope)
	}
}

func TestTokenService_RejectsExpiredToken(t *testing.T) {
	tmpDir := t.TempDir()
	kp, _ := crypto.LoadOrGenerateKeyPair(tmpDir, "RS256", true)

	svc := NewTokenService(TokenServiceConfig{
		PrivateKey: kp.PrivateKey,
		PublicKey:  kp.PublicKey,
		Algorithm:  "RS256",
		KeyID:      "test-key",
		Issuer:     "https://test.saaskit.dev",
		AccessTTL:  -1 * time.Second, // Already expired
	}, nil)

	user := &domain.User{Status: domain.UserStatusActive}
	token, _ := svc.GenerateAccessToken(context.Background(), user, [16]byte{})

	_, err := svc.ValidateAccessToken(token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"user+tag@example.com", true},
		{"invalid", false},
		{"@example.com", false},
		{"", false},
	}

	for _, tt := range tests {
		err := ValidateEmail(tt.email)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateEmail(%q) = %v, want valid=%v", tt.email, err, tt.valid)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		wantErr  error
	}{
		{"short", domain.ErrPasswordTooShort},
		{"12345678", nil},
		{string(make([]byte, 129)), domain.ErrPasswordTooLong},
		{"ValidPassword123!", nil},
	}

	for _, tt := range tests {
		err := ValidatePassword(tt.password)
		if err != tt.wantErr {
			t.Errorf("ValidatePassword(%q) = %v, want %v", tt.password, err, tt.wantErr)
		}
	}
}
