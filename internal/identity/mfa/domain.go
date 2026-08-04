package mfa

import (
	"time"

	"github.com/google/uuid"
)

// MFAType represents the type of MFA method.
type MFAType string

const (
	MFATypeTOTP         MFAType = "totp"
	MFATypeWebAuthn     MFAType = "webauthn"
	MFATypeSMS          MFAType = "sms"
	MFATypeEmail        MFAType = "email"
	MFATypeRecoveryCode MFAType = "recovery_codes"
)

// Method represents an MFA method for a user.
type Method struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Type       MFAType
	Name       string
	SecretEnc  string // Encrypted secret (TOTP secret, etc.)
	Verified   bool
	IsDefault  bool
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

// RecoveryCode represents a single-use recovery code.
type RecoveryCode struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CodeHash  string
	UsedAt    *time.Time
	CreatedAt time.Time
}

// TOTPSetupResponse contains data needed for TOTP enrollment.
type TOTPSetupResponse struct {
	Secret     string   // Unencrypted secret (only returned during setup)
	QRCodeURL  string   // URL for QR code generation
	BackupCodes []string // Recovery codes
}

// VerifyRequest contains data for MFA verification.
type VerifyRequest struct {
	UserID uuid.UUID
	Code   string // TOTP code or recovery code
}
