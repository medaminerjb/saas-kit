package audit

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/saaskit/saaskit/internal/platform/events"
)

func TestNewAuditPublisher(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	pub := NewAuditPublisher(logger, nil)
	if pub == nil {
		t.Fatal("expected non-nil AuditPublisher")
	}
	if pub.logger != logger {
		t.Error("expected logger to match")
	}
	if pub.pool != nil {
		t.Error("expected pool to be nil")
	}
}

func TestAuditPublisher_Publish_NilPool(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	pub := NewAuditPublisher(logger, nil)

	ctx := context.Background()
	ctx = events.WithClientInfo(ctx, "127.0.0.1", "Mozilla/5.0")

	uID := uuid.New()
	event := events.Event{
		Type:      "user.login",
		ActorID:   &uID,
		TargetID:  &uID,
		Timestamp: time.Now(),
		Payload:   map[string]any{"method": "password"},
	}

	err := pub.Publish(ctx, event)
	if err != nil {
		t.Errorf("expected no error when pool is nil, got %v", err)
	}
}
