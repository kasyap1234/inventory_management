package repositories

import (
	"context"
	"fmt"
	"strings"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WebhookSubscriptionRepository interface for webhook subscription operations
type WebhookSubscriptionRepository interface {
	Create(ctx context.Context, webhook *models.WebhookSubscription) error
	Update(ctx context.Context, webhook *models.WebhookSubscription) error
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.WebhookSubscription, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*models.WebhookSubscription, error)
	Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
	GetActiveWebhooksForEvent(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.WebhookSubscription, error)
	UpdateDeliveryStatus(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, success bool) error
}

// webhookSubscriptionRepo implements WebhookSubscriptionRepository
type webhookSubscriptionRepo struct {
	db *pgxpool.Pool
}

// NewWebhookSubscriptionRepo creates a new webhook subscription repository
func NewWebhookSubscriptionRepo(db *pgxpool.Pool) WebhookSubscriptionRepository {
	return &webhookSubscriptionRepo{db: db}
}

// Create creates a new webhook subscription
func (r *webhookSubscriptionRepo) Create(ctx context.Context, webhook *models.WebhookSubscription) error {
	if webhook.ID == uuid.Nil {
		webhook.ID = uuid.New()
	}

	eventTypesJSON, err := webhook.MarshalEventTypes()
	if err != nil {
		return fmt.Errorf("failed to marshal event types: %w", err)
	}

	headersJSON, err := webhook.MarshalHeaders()
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	query := `
		INSERT INTO webhook_subscriptions (
			id, tenant_id, name, url, secret, event_types, headers,
			timeout_seconds, retry_count, is_active, failure_count,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW()
		)`

	_, err = r.db.Exec(ctx, query,
		webhook.ID, webhook.TenantID, webhook.Name, webhook.URL, webhook.Secret,
		eventTypesJSON, headersJSON, webhook.TimeoutSeconds, webhook.RetryCount,
		webhook.IsActive, webhook.FailureCount,
	)

	if err != nil {
		// Check for unique constraint violation (duplicate name)
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "23505") {
			return fmt.Errorf("webhook subscription with name '%s' already exists", webhook.Name)
		}
		return fmt.Errorf("failed to create webhook subscription: %w", err)
	}

	return nil
}

// Update updates an existing webhook subscription
func (r *webhookSubscriptionRepo) Update(ctx context.Context, webhook *models.WebhookSubscription) error {
	eventTypesJSON, err := webhook.MarshalEventTypes()
	if err != nil {
		return fmt.Errorf("failed to marshal event types: %w", err)
	}

	headersJSON, err := webhook.MarshalHeaders()
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	query := `
		UPDATE webhook_subscriptions SET
			name = $3, url = $4, secret = $5, event_types = $6, headers = $7,
			timeout_seconds = $8, retry_count = $9, is_active = $10, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query,
		webhook.TenantID, webhook.ID, webhook.Name, webhook.URL, webhook.Secret,
		eventTypesJSON, headersJSON, webhook.TimeoutSeconds, webhook.RetryCount,
		webhook.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update webhook subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook subscription not found")
	}

	return nil
}

// GetByID retrieves a webhook subscription by ID
func (r *webhookSubscriptionRepo) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.WebhookSubscription, error) {
	query := `
		SELECT id, tenant_id, name, url, secret, event_types, headers,
			   timeout_seconds, retry_count, is_active, last_success_at,
			   last_failure_at, failure_count, created_at, updated_at
		FROM webhook_subscriptions
		WHERE tenant_id = $1 AND id = $2`

	var webhook models.WebhookSubscription
	var eventTypesJSON, headersJSON []byte

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&webhook.ID, &webhook.TenantID, &webhook.Name, &webhook.URL, &webhook.Secret,
		&eventTypesJSON, &headersJSON, &webhook.TimeoutSeconds, &webhook.RetryCount,
		&webhook.IsActive, &webhook.LastSuccessAt, &webhook.LastFailureAt,
		&webhook.FailureCount, &webhook.CreatedAt, &webhook.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("webhook subscription not found")
		}
		return nil, fmt.Errorf("failed to get webhook subscription: %w", err)
	}

	if err := webhook.UnmarshalEventTypes(eventTypesJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event types: %w", err)
	}

	if err := webhook.UnmarshalHeaders(headersJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
	}

	return &webhook, nil
}

// List retrieves all webhook subscriptions for a tenant
func (r *webhookSubscriptionRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*models.WebhookSubscription, error) {
	query := `
		SELECT id, tenant_id, name, url, secret, event_types, headers,
			   timeout_seconds, retry_count, is_active, last_success_at,
			   last_failure_at, failure_count, created_at, updated_at
		FROM webhook_subscriptions
		WHERE tenant_id = $1
		ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhook subscriptions: %w", err)
	}
	defer rows.Close()

	var webhooks []*models.WebhookSubscription

	for rows.Next() {
		var webhook models.WebhookSubscription
		var eventTypesJSON, headersJSON []byte

		err := rows.Scan(
			&webhook.ID, &webhook.TenantID, &webhook.Name, &webhook.URL, &webhook.Secret,
			&eventTypesJSON, &headersJSON, &webhook.TimeoutSeconds, &webhook.RetryCount,
			&webhook.IsActive, &webhook.LastSuccessAt, &webhook.LastFailureAt,
			&webhook.FailureCount, &webhook.CreatedAt, &webhook.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook subscription: %w", err)
		}

		if err := webhook.UnmarshalEventTypes(eventTypesJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event types: %w", err)
		}

		if err := webhook.UnmarshalHeaders(headersJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
		}

		webhooks = append(webhooks, &webhook)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating webhook subscriptions: %w", err)
	}

	return webhooks, nil
}

// Delete deletes a webhook subscription
func (r *webhookSubscriptionRepo) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM webhook_subscriptions WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook subscription not found")
	}

	return nil
}

// GetActiveWebhooksForEvent retrieves active webhooks for a specific event type
func (r *webhookSubscriptionRepo) GetActiveWebhooksForEvent(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.WebhookSubscription, error) {
	query := `
		SELECT id, tenant_id, name, url, secret, event_types, headers,
			   timeout_seconds, retry_count, is_active, last_success_at,
			   last_failure_at, failure_count, created_at, updated_at
		FROM webhook_subscriptions
		WHERE tenant_id = $1 AND is_active = true AND $2 = ANY(event_types)
		ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query, tenantID, eventType)
	if err != nil {
		return nil, fmt.Errorf("failed to get active webhooks for event: %w", err)
	}
	defer rows.Close()

	var webhooks []*models.WebhookSubscription

	for rows.Next() {
		var webhook models.WebhookSubscription
		var eventTypesJSON, headersJSON []byte

		err := rows.Scan(
			&webhook.ID, &webhook.TenantID, &webhook.Name, &webhook.URL, &webhook.Secret,
			&eventTypesJSON, &headersJSON, &webhook.TimeoutSeconds, &webhook.RetryCount,
			&webhook.IsActive, &webhook.LastSuccessAt, &webhook.LastFailureAt,
			&webhook.FailureCount, &webhook.CreatedAt, &webhook.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook subscription: %w", err)
		}

		if err := webhook.UnmarshalEventTypes(eventTypesJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event types: %w", err)
		}

		if err := webhook.UnmarshalHeaders(headersJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
		}

		webhooks = append(webhooks, &webhook)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active webhooks: %w", err)
	}

	return webhooks, nil
}

// UpdateDeliveryStatus updates the delivery status of a webhook
func (r *webhookSubscriptionRepo) UpdateDeliveryStatus(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, success bool) error {
	var query string
	if success {
		query = `
			UPDATE webhook_subscriptions SET
				last_success_at = NOW(),
				failure_count = 0,
				updated_at = NOW()
			WHERE tenant_id = $1 AND id = $2`
	} else {
		query = `
			UPDATE webhook_subscriptions SET
				last_failure_at = NOW(),
				failure_count = failure_count + 1,
				updated_at = NOW()
			WHERE tenant_id = $1 AND id = $2`
	}

	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to update webhook delivery status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook subscription not found")
	}

	return nil
}
