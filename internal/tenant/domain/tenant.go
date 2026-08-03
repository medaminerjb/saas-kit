package domain

import (
	"time"

	"github.com/google/uuid"
)

// TenantStatus represents the lifecycle state of a tenant/organization.
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
)

// Tenant represents an organization in the system.
type Tenant struct {
	ID        uuid.UUID    `json:"id"`
	Name      string       `json:"name"`
	Slug      string       `json:"slug"`
	Status    TenantStatus `json:"status"`
	Plan      string       `json:"plan"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// MemberRole represents the role of a user within a tenant.
type MemberRole string

const (
	RoleOwner  MemberRole = "owner"
	RoleAdmin  MemberRole = "admin"
	RoleMember MemberRole = "member"
)

// Member represents user membership in a tenant.
type Member struct {
	TenantID uuid.UUID  `json:"tenant_id"`
	UserID   uuid.UUID  `json:"user_id"`
	Role     MemberRole `json:"role"`
	JoinedAt time.Time  `json:"joined_at"`

	// User details joined from the users table (optional context)
	Email     string  `json:"email,omitempty"`
	Name      string  `json:"name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// Invitation represents a pending request for a user to join a tenant.
type Invitation struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	Email      string     `json:"email"`
	Role       MemberRole `json:"role"`
	TokenHash  string     `json:"-"`
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// IsExpired returns true if the invitation has expired.
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsAccepted returns true if the invitation was accepted.
func (i *Invitation) IsAccepted() bool {
	return i.AcceptedAt != nil
}
