package domain

import (
	"time"

	"github.com/google/uuid"
)

// TokenType represents the purpose of an identity token.
type TokenType string

const (
	TokenTypePasswordReset      TokenType = "password_reset"
	TokenTypeEmailVerification  TokenType = "email_verification"
	TokenTypeInvite             TokenType = "invite"
	TokenTypeMagicLink          TokenType = "magic_link"
	TokenTypeMFA                TokenType = "mfa"
	TokenTypeDeviceVerification TokenType = "device_verification"
)

// IdentityToken is a one-time-use token for password resets,
// email verification, invitations, magic links, etc.
// The raw token is never stored — only its HMAC hash.
type IdentityToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Type      TokenType  `json:"type"`
	Hash      string     `json:"-"` // HMAC(server_secret, token)
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// IsValid returns true if the token has not been used and is not expired.
func (t *IdentityToken) IsValid() bool {
	return t.UsedAt == nil && time.Now().Before(t.ExpiresAt)
}
