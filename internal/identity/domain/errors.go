package domain

import "errors"

// Domain errors for the identity module.
// These are sentinel errors that services return and handlers map to HTTP status codes.
var (
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExists    = errors.New("user with this email already exists")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrAccountDisabled      = errors.New("account is disabled")
	ErrAccountLocked        = errors.New("account is locked")
	ErrAccountNotVerified   = errors.New("email address not verified")
	ErrInvalidToken         = errors.New("invalid or expired token")
	ErrSessionExpired       = errors.New("session has expired")
	ErrSessionRevoked       = errors.New("session has been revoked")
	ErrPasswordRequired     = errors.New("password is required")
	ErrPasswordTooShort     = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong      = errors.New("password must be at most 128 characters")
	ErrInvalidEmail         = errors.New("invalid email address")
	ErrTokenAlreadyUsed     = errors.New("token has already been used")
	ErrOAuthAccountNotFound = errors.New("no linked account for this provider")
)
