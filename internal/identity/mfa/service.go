package mfa

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/medaminerjb/saas-kit/internal/platform/crypto"
)

var (
	ErrInvalidCode      = errors.New("invalid mfa code")
	ErrAlreadyEnabled   = errors.New("mfa already enabled")
	ErrNoVerifiedMethod = errors.New("no verified mfa method")
)

// Service handles MFA business logic.
type Service struct {
	repo     Repository
	envelope *crypto.Envelope
	logger   *slog.Logger
}

// ServiceConfig holds MFA service configuration.
type ServiceConfig struct {
	Repository Repository
	Envelope   *crypto.Envelope
	Logger     *slog.Logger
}

// NewService creates a new MFA service.
func NewService(cfg ServiceConfig) *Service {
	return &Service{
		repo:     cfg.Repository,
		envelope: cfg.Envelope,
		logger:   cfg.Logger,
	}
}

// SetupTOTP initiates TOTP enrollment for a user.
func (s *Service) SetupTOTP(ctx context.Context, userID uuid.UUID, name string) (*TOTPSetupResponse, error) {
	// Check if TOTP already exists
	methods, err := s.repo.GetMethodsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, m := range methods {
		if m.Type == MFATypeTOTP && m.Verified {
			return nil, ErrAlreadyEnabled
		}
	}

	// Generate TOTP secret
	secret := generateTOTPSecret()

	// Encrypt secret
	secretEnc, err := s.envelope.Encrypt([]byte(secret), "mfa_totp_secret")
	if err != nil {
		return nil, fmt.Errorf("encrypting totp secret: %w", err)
	}

	// Create unverified method
	method := &Method{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      MFATypeTOTP,
		Name:      name,
		SecretEnc: secretEnc,
		Verified:  false,
		IsDefault: len(methods) == 0,
	}

	if err := s.repo.CreateMethod(ctx, method); err != nil {
		return nil, fmt.Errorf("creating totp method: %w", err)
	}

	// Generate QR code URL (otpauth format)
	issuer := "SaaSKit"
	accountName := fmt.Sprintf("%s:%s", issuer, name)
	qrURL := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s", accountName, secret, issuer)

	// Generate recovery codes
	recoveryCodes, err := s.generateRecoveryCodes(ctx, userID)
	if err != nil {
		s.logger.Warn("failed to generate recovery codes", "user_id", userID, "error", err)
		return &TOTPSetupResponse{
			Secret:      secret,
			QRCodeURL:   qrURL,
			BackupCodes: nil,
		}, nil
	}

	return &TOTPSetupResponse{
		Secret:      secret,
		QRCodeURL:   qrURL,
		BackupCodes: recoveryCodes,
	}, nil
}

// VerifyTOTP verifies a TOTP code and marks the method as verified.
func (s *Service) VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	methods, err := s.repo.GetMethodsByUserID(ctx, userID)
	if err != nil {
		return err
	}

	var totpMethod *Method
	for _, m := range methods {
		if m.Type == MFATypeTOTP && !m.Verified {
			totpMethod = m
			break
		}
	}
	if totpMethod == nil {
		return ErrMethodNotFound
	}

	// Decrypt secret
	secretBytes, err := s.envelope.Decrypt(totpMethod.SecretEnc, "mfa_totp_secret")
	if err != nil {
		return fmt.Errorf("decrypting totp secret: %w", err)
	}
	secret := string(secretBytes)

	// Validate TOTP code
	if !validateTOTP(secret, code, time.Now()) {
		return ErrInvalidCode
	}

	// Mark as verified
	totpMethod.Verified = true
	now := time.Now()
	totpMethod.LastUsedAt = &now
	if err := s.repo.UpdateMethod(ctx, totpMethod); err != nil {
		return fmt.Errorf("marking method verified: %w", err)
	}

	return nil
}

// VerifyCode verifies an MFA code (TOTP or recovery code).
func (s *Service) VerifyCode(ctx context.Context, userID uuid.UUID, code string) error {
	// Try TOTP first
	methods, err := s.repo.GetMethodsByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, m := range methods {
		if m.Type == MFATypeTOTP && m.Verified {
			secretBytes, err := s.envelope.Decrypt(m.SecretEnc, "mfa_totp_secret")
			if err != nil {
				continue
			}
			if validateTOTP(string(secretBytes), code, time.Now()) {
				// Update last used
				now := time.Now()
				m.LastUsedAt = &now
				_ = s.repo.UpdateMethod(ctx, m)
				return nil
			}
		}
	}

	// Try recovery codes
	codes, err := s.repo.GetRecoveryCodesByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, rc := range codes {
		if rc.CodeHash == hashRecoveryCode(code) && rc.UsedAt == nil {
			if err := s.repo.MarkRecoveryCodeUsed(ctx, rc.ID, userID); err != nil {
				return err
			}
			return nil
		}
	}

	return ErrInvalidCode
}

// HasEnabledMFA checks if a user has any verified MFA methods.
func (s *Service) HasEnabledMFA(ctx context.Context, userID uuid.UUID) (bool, error) {
	methods, err := s.repo.GetMethodsByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, m := range methods {
		if m.Verified {
			return true, nil
		}
	}
	return false, nil
}

// RemoveMethod removes an MFA method.
func (s *Service) RemoveMethod(ctx context.Context, userID, methodID uuid.UUID) error {
	// Check if this is the only verified method
	methods, err := s.repo.GetMethodsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	verifiedCount := 0
	for _, m := range methods {
		if m.Verified && m.ID != methodID {
			verifiedCount++
		}
	}
	if verifiedCount == 0 {
		return errors.New("cannot remove last verified mfa method")
	}

	return s.repo.DeleteMethod(ctx, methodID, userID)
}

// generateRecoveryCodes generates and stores recovery codes for a user.
func (s *Service) generateRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	const numCodes = 10
	codes := make([]string, numCodes)
	recoveryCodes := make([]*RecoveryCode, numCodes)

	for i := 0; i < numCodes; i++ {
		code := generateRecoveryCode()
		codes[i] = code
		recoveryCodes[i] = &RecoveryCode{
			ID:       uuid.New(),
			UserID:   userID,
			CodeHash: hashRecoveryCode(code),
		}
	}

	if err := s.repo.CreateRecoveryCodes(ctx, recoveryCodes); err != nil {
		return nil, err
	}

	return codes, nil
}

// generateTOTPSecret generates a random 20-byte base32-encoded secret.
func generateTOTPSecret() string {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		panic(err) // Should never happen
	}
	return base32.StdEncoding.EncodeToString(b)
}

// generateRecoveryCode generates a random 8-character recovery code.
func generateRecoveryCode() string {
	const charset = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	b := make([]byte, 8)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// hashRecoveryCode hashes a recovery code for storage.
func hashRecoveryCode(code string) string {
	// Simple hash for now - in production use proper password hashing
	return fmt.Sprintf("%x", code)
}

// validateTOTP validates a TOTP code against a secret using standard TOTP algorithm.
func validateTOTP(secret, code string, t time.Time) bool {
	// Decode base32 secret
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return false
	}

	// Check current time and adjacent time windows (allow 1 step skew)
	for i := -1; i <= 1; i++ {
		if validateTOTPAtTime(key, code, t.Add(time.Duration(i)*30*time.Second)) {
			return true
		}
	}
	return false
}

// validateTOTPAtTime validates TOTP code at a specific time.
func validateTOTPAtTime(key []byte, code string, t time.Time) bool {
	timeSteps := t.Unix() / 30

	// Generate HMAC-SHA1
	h := hmac.New(sha1.New, key)
	h.Write([]byte(fmt.Sprintf("%016x", timeSteps)))
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := int(hash[len(hash)-1] & 0x0f)
	binary := ((int(hash[offset]) & 0x7f) << 24) |
		((int(hash[offset+1]) & 0xff) << 16) |
		((int(hash[offset+2]) & 0xff) << 8) |
		(int(hash[offset+3]) & 0xff)

	otp := binary % 1000000

	// Format as 6-digit string with leading zeros
	expectedCode := fmt.Sprintf("%06d", otp)
	return code == expectedCode
}
