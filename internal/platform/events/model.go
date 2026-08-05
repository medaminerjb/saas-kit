package events

import (
	"github.com/google/uuid"
)

// EventType represents the type of event.
type EventType string

const (
	// User events
	EventUserCreated       EventType = "user.created"
	EventUserUpdated       EventType = "user.updated"
	EventUserDeleted       EventType = "user.deleted"
	EventUserDisabled      EventType = "user.disabled"
	EventUserPasswordReset EventType = "user.password_reset"

	// Tenant events
	EventTenantCreated EventType = "tenant.created"
	EventTenantUpdated EventType = "tenant.updated"
	EventTenantDeleted EventType = "tenant.deleted"
	EventMemberInvited EventType = "member.invited"
	EventMemberJoined  EventType = "member.joined"
	EventMemberRemoved EventType = "member.removed"
	EventMemberUpdated EventType = "member.updated"

	// API Key events
	EventAPIKeyCreated EventType = "api_key.created"
	EventAPIKeyRevoked EventType = "api_key.revoked"
	EventAPIKeyDeleted EventType = "api_key.deleted"

	// Session events
	EventSessionCreated EventType = "session.created"
	EventSessionRevoked EventType = "session.revoked"
)

// EventDataAPIKeyCreated represents data for api_key.created event.
type EventDataAPIKeyCreated struct {
	APIKeyID  uuid.UUID `json:"api_key_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Scopes    []string  `json:"scopes"`
	CreatedBy uuid.UUID `json:"created_by"`
}
