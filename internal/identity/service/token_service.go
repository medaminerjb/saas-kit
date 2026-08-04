// Package service implements the identity business logic.
package service

import (
	"context"
	"crypto"
	"fmt"
	"log/slog"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/medaminerjb/saas-kit/internal/identity/domain"
)

// JWTClaims are the claims included in SaaSKit access tokens.
type JWTClaims struct {
	jwt.RegisteredClaims
	SessionID   string   `json:"sid,omitempty"`
	TenantID    *string  `json:"tenant,omitempty"`
	TenantRole  *string  `json:"tenant_role,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	MFAVerified bool     `json:"mfa_verified,omitempty"`
	Scope       string   `json:"scope,omitempty"`
}

// TokenService handles JWT access token generation and validation.
type TokenService struct {
	privateKey crypto.Signer
	publicKey  crypto.PublicKey
	algorithm  string
	keyID      string
	issuer     string
	accessTTL  time.Duration
	logger     *slog.Logger
}

// TokenServiceConfig holds TokenService configuration.
type TokenServiceConfig struct {
	PrivateKey crypto.Signer
	PublicKey  crypto.PublicKey
	Algorithm  string
	KeyID      string
	Issuer     string
	AccessTTL  time.Duration
}

// NewTokenService creates a new JWT token service.
func NewTokenService(cfg TokenServiceConfig, logger *slog.Logger) *TokenService {
	return &TokenService{
		privateKey: cfg.PrivateKey,
		publicKey:  cfg.PublicKey,
		algorithm:  cfg.Algorithm,
		keyID:      cfg.KeyID,
		issuer:     cfg.Issuer,
		accessTTL:  cfg.AccessTTL,
		logger:     logger,
	}
}

// GenerateAccessToken creates a signed JWT access token for the given user and session.
// Optional tenantRole and permissions can be provided for RBAC.
func (s *TokenService) GenerateAccessToken(_ context.Context, user *domain.User, sessionID uuid.UUID, tenantRole *string, permissions []string, mfaVerified bool) (string, error) {
	now := time.Now()
	jti := uuid.New().String()

	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			ID:        jti,
		},
		SessionID:   sessionID.String(),
		TenantRole:  tenantRole,
		Permissions: permissions,
		MFAVerified: mfaVerified,
		Scope:       "openid profile email",
	}

	if user.TenantID != nil {
		tid := user.TenantID.String()
		claims.TenantID = &tid
	}

	signingMethod := signingMethodFromAlgorithm(s.algorithm)
	token := jwt.NewWithClaims(signingMethod, claims)
	token.Header["kid"] = s.keyID

	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("signing access token: %w", err)
	}

	return signed, nil
}

// ValidateAccessToken parses and validates a JWT access token.
func (s *TokenService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return s.publicKey, nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, keyFunc,
		jwt.WithIssuer(s.issuer),
		jwt.WithValidMethods([]string{s.algorithm}),
	)
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func signingMethodFromAlgorithm(alg string) jwt.SigningMethod {
	switch alg {
	case "ES256":
		return jwt.SigningMethodES256
	case "EdDSA":
		return jwt.SigningMethodEdDSA
	default:
		return jwt.SigningMethodRS256
	}
}

// ValidateEmail checks if an email address is syntactically valid.
func ValidateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return domain.ErrInvalidEmail
	}
	return nil
}

// ValidatePassword checks password strength requirements.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return domain.ErrPasswordTooShort
	}
	if len(password) > 128 {
		return domain.ErrPasswordTooLong
	}
	return nil
}
