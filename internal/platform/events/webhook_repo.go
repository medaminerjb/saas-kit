package events

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medaminerjb/saas-kit/internal/sqlcgen"
)

// WebhookRepository defines the interface for webhook persistence.
type WebhookRepository interface {
	CreateSubscription(ctx context.Context, sub *WebhookSubscription) error
	GetSubscriptionByID(ctx context.Context, id, tenantID uuid.UUID) (*WebhookSubscription, error)
	ListSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*WebhookSubscription, error)
	ListActiveSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*WebhookSubscription, error)
	UpdateSubscription(ctx context.Context, sub *WebhookSubscription) error
	DeleteSubscription(ctx context.Context, id, tenantID uuid.UUID) error
	CreateDelivery(ctx context.Context, delivery *WebhookDelivery) error
	ListFailedDeliveries(ctx context.Context) ([]*WebhookDelivery, error)
}

// WebhookRepo implements WebhookRepository using sqlc-generated queries.
type WebhookRepo struct {
	pool    *pgxpool.Pool
	queries *sqlcgen.Queries
}

// NewWebhookRepo creates a new WebhookRepo.
func NewWebhookRepo(pool *pgxpool.Pool) *WebhookRepo {
	return &WebhookRepo{
		pool:    pool,
		queries: sqlcgen.New(pool),
	}
}

func (r *WebhookRepo) CreateSubscription(ctx context.Context, sub *WebhookSubscription) error {
	params := sqlcgen.CreateWebhookSubscriptionParams{
		TenantID:   sub.TenantID,
		Url:        sub.URL,
		Secret:     sub.Secret,
		EventTypes: eventTypeSlice(sub.EventTypes),
		Status:     string(sub.Status),
	}

	result, err := r.queries.CreateWebhookSubscription(ctx, params)
	if err != nil {
		return err
	}

	*sub = r.sqlcToSubscription(result)
	return nil
}

func (r *WebhookRepo) GetSubscriptionByID(ctx context.Context, id, tenantID uuid.UUID) (*WebhookSubscription, error) {
	result, err := r.queries.GetWebhookSubscriptionByID(ctx, sqlcgen.GetWebhookSubscriptionByIDParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}

	sub := r.sqlcToSubscription(result)
	return &sub, nil
}

func (r *WebhookRepo) ListSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*WebhookSubscription, error) {
	results, err := r.queries.ListWebhookSubscriptionsByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	subs := make([]*WebhookSubscription, len(results))
	for i, result := range results {
		sub := r.sqlcToSubscription(result)
		subs[i] = &sub
	}

	return subs, nil
}

func (r *WebhookRepo) ListActiveSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*WebhookSubscription, error) {
	results, err := r.queries.ListActiveWebhookSubscriptionsByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	subs := make([]*WebhookSubscription, len(results))
	for i, result := range results {
		sub := r.sqlcToSubscription(result)
		subs[i] = &sub
	}

	return subs, nil
}

func (r *WebhookRepo) UpdateSubscription(ctx context.Context, sub *WebhookSubscription) error {
	params := sqlcgen.UpdateWebhookSubscriptionParams{
		ID:         sub.ID,
		TenantID:   sub.TenantID,
		Url:        sub.URL,
		Secret:     sub.Secret,
		EventTypes: eventTypeSlice(sub.EventTypes),
		Status:     string(sub.Status),
	}

	result, err := r.queries.UpdateWebhookSubscription(ctx, params)
	if err != nil {
		return err
	}

	*sub = r.sqlcToSubscription(result)
	return nil
}

func (r *WebhookRepo) DeleteSubscription(ctx context.Context, id, tenantID uuid.UUID) error {
	return r.queries.DeleteWebhookSubscription(ctx, sqlcgen.DeleteWebhookSubscriptionParams{
		ID:       id,
		TenantID: tenantID,
	})
}

func (r *WebhookRepo) CreateDelivery(ctx context.Context, delivery *WebhookDelivery) error {
	var statusCode *int32
	if delivery.StatusCode != nil {
		sc := int32(*delivery.StatusCode)
		statusCode = &sc
	}

	params := sqlcgen.CreateWebhookDeliveryParams{
		SubscriptionID: delivery.SubscriptionID,
		EventID:        delivery.EventID,
		EventType:      string(delivery.EventType),
		StatusCode:     statusCode,
		ErrorMessage:   delivery.ErrorMessage,
		Success:        delivery.Success,
	}

	result, err := r.queries.CreateWebhookDelivery(ctx, params)
	if err != nil {
		return err
	}

	*delivery = r.sqlcToDelivery(result)
	return nil
}

func (r *WebhookRepo) ListFailedDeliveries(ctx context.Context) ([]*WebhookDelivery, error) {
	results, err := r.queries.ListFailedWebhookDeliveries(ctx)
	if err != nil {
		return nil, err
	}

	deliveries := make([]*WebhookDelivery, len(results))
	for i, result := range results {
		delivery := r.sqlcToDelivery(result)
		deliveries[i] = &delivery
	}

	return deliveries, nil
}

func (r *WebhookRepo) sqlcToSubscription(result sqlcgen.WebhookSubscription) WebhookSubscription {
	var lastUsedAt *time.Time
	if result.LastUsedAt.Valid {
		lastUsedAt = &result.LastUsedAt.Time
	}

	return WebhookSubscription{
		ID:         result.ID,
		TenantID:   result.TenantID,
		URL:        result.Url,
		Secret:     result.Secret,
		EventTypes: eventTypesFromSlice(result.EventTypes),
		Status:     WebhookStatus(result.Status),
		CreatedAt:  result.CreatedAt,
		UpdatedAt:  result.UpdatedAt,
		LastUsedAt: lastUsedAt,
	}
}

func (r *WebhookRepo) sqlcToDelivery(result sqlcgen.WebhookDelivery) WebhookDelivery {
	var statusCode *int
	if result.StatusCode != nil {
		sc := int(*result.StatusCode)
		statusCode = &sc
	}

	return WebhookDelivery{
		ID:             result.ID,
		SubscriptionID: result.SubscriptionID,
		EventID:        result.EventID,
		EventType:      EventType(result.EventType),
		StatusCode:     statusCode,
		ErrorMessage:   result.ErrorMessage,
		Success:        result.Success,
		AttemptedAt:    result.AttemptedAt,
	}
}

func eventTypeSlice(types []EventType) []string {
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = string(t)
	}
	return result
}

func eventTypesFromSlice(types []string) []EventType {
	result := make([]EventType, len(types))
	for i, t := range types {
		result[i] = EventType(t)
	}
	return result
}
