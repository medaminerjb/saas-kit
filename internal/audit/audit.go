// Package audit provides append-only audit logging for identity operations.
package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medaminerjb/saas-kit/internal/platform/events"
	"github.com/medaminerjb/saas-kit/internal/sqlcgen"
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
	pool   *pgxpool.Pool
}

// NewAuditPublisher creates a publisher that records events as audit logs.
func NewAuditPublisher(logger *slog.Logger, pool *pgxpool.Pool) *AuditPublisher {
	return &AuditPublisher{
		logger: logger,
		pool:   pool,
	}
}

// Publish records an event as an audit log entry.
func (p *AuditPublisher) Publish(ctx context.Context, event events.Event) error {
	p.logger.InfoContext(ctx, "audit",
		slog.String("event", event.Type),
		slog.Time("timestamp", event.Timestamp),
	)

	if p.pool == nil {
		return nil
	}

	// Map events.Event to sqlcgen.CreateAuditLogParams
	var tenantID pgtype.UUID
	if event.TenantID != nil {
		tenantID = pgtype.UUID{Bytes: *event.TenantID, Valid: true}
	}

	var actorID pgtype.UUID
	if event.ActorID != nil {
		actorID = pgtype.UUID{Bytes: *event.ActorID, Valid: true}
	}

	var targetID pgtype.UUID
	if event.TargetID != nil {
		targetID = pgtype.UUID{Bytes: *event.TargetID, Valid: true}
	}

	// Extract Client IP and User Agent from context
	ipStr, userAgentStr := events.GetClientInfo(ctx)

	var ipAddress *netip.Addr
	if ipStr != "" {
		if ip, err := netip.ParseAddr(ipStr); err == nil {
			ipAddress = &ip
		}
	}

	var userAgent *string
	if userAgentStr != "" {
		userAgent = &userAgentStr
	}

	// Serialize metadata/payload to JSON bytes
	var metadata []byte
	if event.Payload != nil {
		if mBytes, err := json.Marshal(event.Payload); err == nil {
			metadata = mBytes
		}
	}

	queries := sqlcgen.New(p.pool)
	err := queries.CreateAuditLog(ctx, sqlcgen.CreateAuditLogParams{
		TenantID:  tenantID,
		ActorID:   actorID,
		TargetID:  targetID,
		Event:     sqlcgen.AuditEventType(event.Type),
		IpAddress: ipAddress,
		UserAgent: userAgent,
		Metadata:  metadata,
	})
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to persist audit log",
			slog.String("event", event.Type),
			slog.Any("error", err),
		)
		return err
	}

	return nil
}
