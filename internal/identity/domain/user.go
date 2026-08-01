// Package domain defines the core identity domain models.
// These types have zero external dependencies and represent the
// canonical entities of the identity system.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the lifecycle state of a user account.
type UserStatus string

const (
	UserStatusActive              UserStatus = "active"
	UserStatusDisabled            UserStatus = "disabled"
	UserStatusLocked              UserStatus = "locked"
	UserStatusPendingVerification UserStatus = "pending_verification"
	UserStatusInvited             UserStatus = "invited"
	UserStatusDeleted             UserStatus = "deleted"
)

// User represents an identity in the system.
type User struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      *uuid.UUID `json:"tenant_id,omitempty"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	PasswordHash  *string    `json:"-"` // Never serialized
	Status        UserStatus `json:"status"`
	EmailVerified bool       `json:"email_verified"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	DeletedAt     *time.Time `json:"-"`
}

// IsActive returns true if the user can authenticate.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// CanLogin returns true if the user is in a state that allows authentication.
func (u *User) CanLogin() bool {
	return u.Status == UserStatusActive || u.Status == UserStatusPendingVerification
}

// HasPassword returns true if the user has a password set (not OAuth-only).
func (u *User) HasPassword() bool {
	return u.PasswordHash != nil && *u.PasswordHash != ""
}
