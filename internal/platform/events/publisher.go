// Package events defines the event publishing interface and default implementations.
// All identity operations emit events, enabling audit logging, webhooks, and
// future integration with message brokers (Kafka, NATS, Redis Streams).
package events

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event emitted by the system.
type Event struct {
	Type      string     `json:"type"`
	TenantID  *uuid.UUID `json:"tenant_id,omitempty"`
	ActorID   *uuid.UUID `json:"actor_id,omitempty"`
	TargetID  *uuid.UUID `json:"target_id,omitempty"`
	Payload   any        `json:"payload,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// Publisher is the interface for publishing domain events.
type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

// LogPublisher logs events using slog. Default implementation for development.
type LogPublisher struct {
	logger *slog.Logger
}

// NewLogPublisher creates a publisher that logs events via slog.
func NewLogPublisher(logger *slog.Logger) *LogPublisher {
	return &LogPublisher{logger: logger}
}

// Publish logs the event at info level.
func (p *LogPublisher) Publish(ctx context.Context, event Event) error {
	attrs := []any{
		slog.String("event_type", event.Type),
		slog.Time("timestamp", event.Timestamp),
	}
	if event.TenantID != nil {
		attrs = append(attrs, slog.String("tenant_id", event.TenantID.String()))
	}
	if event.ActorID != nil {
		attrs = append(attrs, slog.String("actor_id", event.ActorID.String()))
	}
	if event.TargetID != nil {
		attrs = append(attrs, slog.String("target_id", event.TargetID.String()))
	}
	p.logger.InfoContext(ctx, "event published", attrs...)
	return nil
}

// MultiPublisher fans out events to multiple publishers.
type MultiPublisher struct {
	publishers []Publisher
}

// NewMultiPublisher creates a publisher that sends to all provided publishers.
func NewMultiPublisher(publishers ...Publisher) *MultiPublisher {
	return &MultiPublisher{publishers: publishers}
}

// Publish sends the event to all publishers. Returns the first error encountered.
func (m *MultiPublisher) Publish(ctx context.Context, event Event) error {
	for _, pub := range m.publishers {
		if err := pub.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// NoopPublisher discards all events. Useful for testing.
type NoopPublisher struct{}

// Publish does nothing.
func (n *NoopPublisher) Publish(_ context.Context, _ Event) error {
	return nil
}
