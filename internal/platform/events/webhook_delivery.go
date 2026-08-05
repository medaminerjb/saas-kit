package events

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// WebhookDeliverer handles webhook delivery to subscribed endpoints.
type WebhookDeliverer struct {
	repo   WebhookRepository
	client *http.Client
	config WebhookDeliveryConfig
	logger *slog.Logger
}

// NewWebhookDeliverer creates a new WebhookDeliverer.
func NewWebhookDeliverer(repo WebhookRepository, config WebhookDeliveryConfig, logger *slog.Logger) *WebhookDeliverer {
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 5 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.SignatureHeader == "" {
		config.SignatureHeader = "X-SaaSKit-Signature"
	}

	return &WebhookDeliverer{
		repo:   repo,
		client: &http.Client{Timeout: config.Timeout},
		config: config,
		logger: logger,
	}
}

// DeliverEvent delivers an event to all active webhook subscriptions for the tenant.
func (d *WebhookDeliverer) DeliverEvent(ctx context.Context, tenantID uuid.UUID, event Event) error {
	subs, err := d.repo.ListActiveSubscriptionsByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to list webhook subscriptions: %w", err)
	}

	// Filter subscriptions that are interested in this event type
	for _, sub := range subs {
		if d.shouldDeliver(sub, event.Type) {
			go d.deliverToSubscription(ctx, sub, event)
		}
	}

	return nil
}

// shouldDeliver checks if a subscription should receive the event.
func (d *WebhookDeliverer) shouldDeliver(sub *WebhookSubscription, eventType string) bool {
	for _, t := range sub.EventTypes {
		if string(t) == eventType {
			return true
		}
	}
	return false
}

// deliverToSubscription delivers an event to a single subscription with retries.
func (d *WebhookDeliverer) deliverToSubscription(ctx context.Context, sub *WebhookSubscription, event Event) {
	var lastErr error

	for attempt := 0; attempt <= d.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(d.config.RetryDelay):
			case <-ctx.Done():
				return
			}
		}

		err := d.attemptDelivery(ctx, sub, event)
		if err == nil {
			// Success
			d.logger.InfoContext(ctx, "webhook delivered successfully",
				"subscription_id", sub.ID,
				"event_type", event.Type,
				"attempt", attempt+1,
			)
			return
		}

		lastErr = err
		d.logger.WarnContext(ctx, "webhook delivery attempt failed",
			"subscription_id", sub.ID,
			"event_type", event.Type,
			"attempt", attempt+1,
			"error", err,
		)
	}

	// All attempts failed - record failed delivery
	delivery := &WebhookDelivery{
		SubscriptionID: sub.ID,
		EventID:        uuid.New(), // Event ID should come from event if available
		EventType:      EventType(event.Type),
		StatusCode:     nil,
		ErrorMessage:   strPtr(lastErr.Error()),
		Success:        false,
		AttemptedAt:    time.Now().UTC(),
	}

	if err := d.repo.CreateDelivery(context.Background(), delivery); err != nil {
		d.logger.ErrorContext(ctx, "failed to record failed webhook delivery",
			"subscription_id", sub.ID,
			"error", err,
		)
	}
}

// attemptDelivery makes a single HTTP attempt to deliver the webhook.
func (d *WebhookDeliverer) attemptDelivery(ctx context.Context, sub *WebhookSubscription, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sub.URL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SaaSKit-Webhook/1.0")

	// Add signature if secret is configured
	if sub.Secret != nil {
		signature := d.signPayload(payload, *sub.Secret)
		req.Header.Set(d.config.SignatureHeader, signature)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			d.logger.Warn("failed to close response body", "error", err)
		}
	}()

	body, _ := io.ReadAll(resp.Body)

	// Record successful delivery
	delivery := &WebhookDelivery{
		SubscriptionID: sub.ID,
		EventID:        uuid.New(),
		EventType:      EventType(event.Type),
		StatusCode:     intPtr(resp.StatusCode),
		ErrorMessage:   nil,
		Success:        resp.StatusCode >= 200 && resp.StatusCode < 300,
		AttemptedAt:    time.Now().UTC(),
	}

	if err := d.repo.CreateDelivery(context.Background(), delivery); err != nil {
		d.logger.WarnContext(ctx, "failed to record webhook delivery",
			"subscription_id", sub.ID,
			"error", err,
		)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("webhook endpoint returned status %d: %s", resp.StatusCode, string(body))
}

// signPayload creates an HMAC-SHA256 signature for the payload.
func (d *WebhookDeliverer) signPayload(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

// RetryFailedDeliveries retries failed webhook deliveries.
func (d *WebhookDeliverer) RetryFailedDeliveries(ctx context.Context) error {
	deliveries, err := d.repo.ListFailedDeliveries(ctx)
	if err != nil {
		return fmt.Errorf("failed to list failed deliveries: %w", err)
	}

	for _, delivery := range deliveries {
		sub, err := d.repo.GetSubscriptionByID(ctx, delivery.SubscriptionID, uuid.UUID{}) // Need tenant ID
		if err != nil {
			d.logger.WarnContext(ctx, "failed to get subscription for retry",
				"delivery_id", delivery.ID,
				"error", err,
			)
			continue
		}

		// Reconstruct event from delivery (simplified - in production, store full event)
		event := Event{
			Type:      string(delivery.EventType),
			TenantID:  &sub.TenantID,
			ActorID:   nil,
			TargetID:  nil,
			Payload:   nil,
			Timestamp: delivery.AttemptedAt,
		}

		go d.deliverToSubscription(ctx, sub, event)
	}

	return nil
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
