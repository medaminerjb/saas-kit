// Package audit provides append-only audit logging for identity operations.
package audit

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/saaskit/saaskit/internal/platform/events"
)

// EventType defines the type of auditable event.
type EventType string

// All audit event types supported by the identity module.
const (
	EventUserRegistered     EventType = "user.registered"
	EventUserLogin          EventType = "user.login"
	EventUserLogout         EventType = "user.logout"
	EventPasswordChanged    EventType = "user.password_changed"
	EventPasswordReset      EventType = "user.password_reset_requested"
	EventEmailVerified      EventType = "user.email_verified"
	EventUserDisabled       EventType = "user.disabled"
	EventUserDeleted        EventType = "user.deleted"
	EventUserUpdated        EventType = "user.updated"
	EventSessionCreated     EventType = "session.created"
	EventSessionRevoked     EventType = "session.revoked"
	EventTokenRevoked       EventType = "token.revoked"
	EventOIDCClientCreated  EventType = "oidc_client.created"
	EventOIDCClientUpdated  EventType = "oidc_client.updated"
	EventOIDCClientDeleted  EventType = "oidc_client.deleted"
	EventIDPCreated         EventType = "idp.created"
	EventIDPUpdated         EventType = "idp.updated"
)

// Entry is an audit log entry.
type Entry struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  *uuid.UUID `json:"tenant_id,omitempty"`
	ActorID   *uuid.UUID `json:"actor_id,omitempty"`
	TargetID  *uuid.UUID `json:"target_id,omitempty"`
	Event     EventType  `json:"event"`
	IPAddress string     `json:"ip_address,omitempty"`
	UserAgent string     `json:"user_agent,omitempty"`
	Metadata  any        `json:"metadata,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// Service records audit log entries.
type Service struct {
	logger *slog.Logger
	// In Phase 1, audit logs are written via the event publisher.
	// The AuditPublisher below subscribes to events and persists them.
}

// NewService creates a new audit service.
func NewService(logger *slog.Logger) *Service {
	return &Service{logger: logger}
}

// AuditPublisher is an events.Publisher that persists events as audit log entries.
// It wraps the repository layer and can be added to a MultiPublisher.
type AuditPublisher struct {
	logger *slog.Logger
	// store AuditRepository — will be wired when Postgres repo is ready
}

// NewAuditPublisher creates a publisher that records events as audit logs.
func NewAuditPublisher(logger *slog.Logger) *AuditPublisher {
	return &AuditPublisher{logger: logger}
}

// Publish records an event as an audit log entry.
func (p *AuditPublisher) Publish(ctx context.Context, event events.Event) error {
	p.logger.InfoContext(ctx, "audit",
		slog.String("event", event.Type),
		slog.Time("timestamp", event.Timestamp),
	)
	// TODO: persist to audit_logs table via repository
	return nil
}
