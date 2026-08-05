package events

import (
	"time"

	"github.com/google/uuid"
)

// WebhookStatus represents the status of a webhook subscription.
type WebhookStatus string

const (
	WebhookStatusActive   WebhookStatus = "active"
	WebhookStatusInactive WebhookStatus = "inactive"
)

// WebhookSubscription represents a webhook subscription for event delivery.
type WebhookSubscription struct {
	ID          uuid.UUID      `json:"id"`
	TenantID    uuid.UUID      `json:"tenant_id"`
	URL         string         `json:"url"`
	Secret      *string        `json:"secret,omitempty"` // HMAC secret for signature verification
	EventTypes  []EventType    `json:"event_types"`
	Status      WebhookStatus  `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	LastUsedAt  *time.Time     `json:"last_used_at,omitempty"`
}

// WebhookDelivery represents a webhook delivery attempt.
type WebhookDelivery struct {
	ID             uuid.UUID  `json:"id"`
	SubscriptionID uuid.UUID  `json:"subscription_id"`
	EventID        uuid.UUID  `json:"event_id"`
	EventType      EventType `json:"event_type"`
	StatusCode     *int       `json:"status_code,omitempty"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
	Success        bool       `json:"success"`
	AttemptedAt    time.Time  `json:"attempted_at"`
}

// WebhookDeliveryConfig holds configuration for webhook delivery.
type WebhookDeliveryConfig struct {
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	Timeout         time.Duration `json:"timeout"`
	SignatureHeader string        `json:"signature_header"`
}
