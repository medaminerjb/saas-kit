package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated user session.
// The refresh token is never stored — only its HMAC hash.
type Session struct {
	ID                       uuid.UUID  `json:"id"`
	UserID                   uuid.UUID  `json:"user_id"`
	TenantID                 *uuid.UUID `json:"tenant_id,omitempty"`
	RefreshTokenHash         string     `json:"-"`                    // HMAC(server_secret, token)
	PreviousRefreshTokenHash *string    `json:"-"`                    // Previous token hash for grace period
	RotatedAt                *time.Time `json:"rotated_at,omitempty"` // When token was rotated
	GraceRefreshTokenEnc     *string    `json:"-"`                    // Encrypted grace token (optional)
	UserAgent                string     `json:"user_agent,omitempty"`
	IPAddress                string     `json:"ip_address,omitempty"`
	ExpiresAt                time.Time  `json:"expires_at"`
	CreatedAt                time.Time  `json:"created_at"`
	RevokedAt                *time.Time `json:"revoked_at,omitempty"`
}

// IsValid returns true if the session is not expired and not revoked.
func (s *Session) IsValid() bool {
	return s.RevokedAt == nil && time.Now().Before(s.ExpiresAt)
}
