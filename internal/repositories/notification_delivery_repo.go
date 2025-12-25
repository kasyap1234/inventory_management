package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DeliveryFilters represents filters for querying notification deliveries (repository-level)
type DeliveryFilters struct {
	NotificationID *uuid.UUID               `json:"notification_id,omitempty"`
	Channel        *models.NotificationType `json:"channel,omitempty"`
	Status         *string                  `json:"status,omitempty"`
	StartDate      *time.Time               `json:"start_date,omitempty"`
	EndDate        *time.Time               `json:"end_date,omitempty"`
	Limit          int                      `json:"limit,omitempty"`
	Offset         int                      `json:"offset,omitempty"`
}

// notificationDeliveryRepo implements NotificationDeliveryRepository
type notificationDeliveryRepo struct {
	db *pgxpool.Pool
}

// NewNotificationDeliveryRepo creates a new notification delivery repository
func NewNotificationDeliveryRepo(db *pgxpool.Pool) *notificationDeliveryRepo {
	return &notificationDeliveryRepo{db: db}
}

// Create creates a new notification delivery record
func (r *notificationDeliveryRepo) Create(ctx context.Context, delivery *models.NotificationDelivery) error {
	if delivery.ID == uuid.Nil {
		delivery.ID = uuid.New()
	}

	responseDataJSON, err := json.Marshal(delivery.ResponseData)
	if err != nil {
		return fmt.Errorf("failed to marshal response data: %w", err)
	}

	query := `
		INSERT INTO notification_deliveries (
			id, tenant_id, notification_id, template_id, webhook_id,
			channel, recipient, status, attempt_count, last_attempt_at,
			delivered_at, error_message, response_data, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW()
		)`

	_, err = r.db.Exec(ctx, query,
		delivery.ID, delivery.TenantID, delivery.NotificationID, delivery.TemplateID,
		delivery.WebhookID, string(delivery.Channel), delivery.Recipient, delivery.Status,
		delivery.AttemptCount, delivery.LastAttemptAt, delivery.DeliveredAt,
		delivery.ErrorMessage, responseDataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification delivery: %w", err)
	}

	return nil
}

// Update updates an existing notification delivery record
func (r *notificationDeliveryRepo) Update(ctx context.Context, delivery *models.NotificationDelivery) error {
	responseDataJSON, err := json.Marshal(delivery.ResponseData)
	if err != nil {
		return fmt.Errorf("failed to marshal response data: %w", err)
	}

	query := `
		UPDATE notification_deliveries SET
			status = $3, attempt_count = $4, last_attempt_at = $5,
			delivered_at = $6, error_message = $7, response_data = $8,
			updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query,
		delivery.TenantID, delivery.ID, delivery.Status, delivery.AttemptCount,
		delivery.LastAttemptAt, delivery.DeliveredAt, delivery.ErrorMessage,
		responseDataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification delivery: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification delivery not found")
	}

	return nil
}

// GetByID retrieves a notification delivery by ID
func (r *notificationDeliveryRepo) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.NotificationDelivery, error) {
	query := `
		SELECT id, tenant_id, notification_id, template_id, webhook_id,
			   channel, recipient, status, attempt_count, last_attempt_at,
			   delivered_at, error_message, response_data, created_at, updated_at
		FROM notification_deliveries
		WHERE tenant_id = $1 AND id = $2`

	var delivery models.NotificationDelivery
	var responseDataJSON []byte

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&delivery.ID, &delivery.TenantID, &delivery.NotificationID, &delivery.TemplateID,
		&delivery.WebhookID, &delivery.Channel, &delivery.Recipient, &delivery.Status,
		&delivery.AttemptCount, &delivery.LastAttemptAt, &delivery.DeliveredAt,
		&delivery.ErrorMessage, &responseDataJSON, &delivery.CreatedAt, &delivery.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("notification delivery not found")
		}
		return nil, fmt.Errorf("failed to get notification delivery: %w", err)
	}

	if len(responseDataJSON) > 0 {
		if err := json.Unmarshal(responseDataJSON, &delivery.ResponseData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response data: %w", err)
		}
	}

	return &delivery, nil
}

// List retrieves notification deliveries with filters
func (r *notificationDeliveryRepo) List(ctx context.Context, tenantID uuid.UUID, filters DeliveryFilters) ([]*models.NotificationDelivery, error) {
	query := `
		SELECT id, tenant_id, notification_id, template_id, webhook_id,
			   channel, recipient, status, attempt_count, last_attempt_at,
			   delivered_at, error_message, response_data, created_at, updated_at
		FROM notification_deliveries
		WHERE tenant_id = $1`

	args := []interface{}{tenantID}
	argIndex := 2

	// Add filters
	if filters.NotificationID != nil {
		query += fmt.Sprintf(" AND notification_id = $%d", argIndex)
		args = append(args, *filters.NotificationID)
		argIndex++
	}

	if filters.Channel != nil {
		query += fmt.Sprintf(" AND channel = $%d", argIndex)
		args = append(args, string(*filters.Channel))
		argIndex++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filters.EndDate)
		argIndex++
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	limit := filters.Limit
	if limit <= 0 {
		limit = 100 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Maximum limit
	}

	query += fmt.Sprintf(" LIMIT $%d", argIndex)
	args = append(args, limit)
	argIndex++

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notification deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*models.NotificationDelivery

	for rows.Next() {
		var delivery models.NotificationDelivery
		var responseDataJSON []byte

		err := rows.Scan(
			&delivery.ID, &delivery.TenantID, &delivery.NotificationID, &delivery.TemplateID,
			&delivery.WebhookID, &delivery.Channel, &delivery.Recipient, &delivery.Status,
			&delivery.AttemptCount, &delivery.LastAttemptAt, &delivery.DeliveredAt,
			&delivery.ErrorMessage, &responseDataJSON, &delivery.CreatedAt, &delivery.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan notification delivery: %w", err)
		}

		if len(responseDataJSON) > 0 {
			if err := json.Unmarshal(responseDataJSON, &delivery.ResponseData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response data: %w", err)
			}
		}

		deliveries = append(deliveries, &delivery)
	}

	return deliveries, nil
}

// GetFailedDeliveries retrieves failed deliveries that can be retried
func (r *notificationDeliveryRepo) GetFailedDeliveries(ctx context.Context, tenantID uuid.UUID, maxRetries int) ([]*models.NotificationDelivery, error) {
	query := `
		SELECT id, tenant_id, notification_id, template_id, webhook_id,
			   channel, recipient, status, attempt_count, last_attempt_at,
			   delivered_at, error_message, response_data, created_at, updated_at
		FROM notification_deliveries
		WHERE tenant_id = $1 AND status = 'failed' AND attempt_count < $2
		ORDER BY created_at ASC
		LIMIT 100`

	rows, err := r.db.Query(ctx, query, tenantID, maxRetries)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*models.NotificationDelivery

	for rows.Next() {
		var delivery models.NotificationDelivery
		var responseDataJSON []byte

		err := rows.Scan(
			&delivery.ID, &delivery.TenantID, &delivery.NotificationID, &delivery.TemplateID,
			&delivery.WebhookID, &delivery.Channel, &delivery.Recipient, &delivery.Status,
			&delivery.AttemptCount, &delivery.LastAttemptAt, &delivery.DeliveredAt,
			&delivery.ErrorMessage, &responseDataJSON, &delivery.CreatedAt, &delivery.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan failed delivery: %w", err)
		}

		if len(responseDataJSON) > 0 {
			if err := json.Unmarshal(responseDataJSON, &delivery.ResponseData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response data: %w", err)
			}
		}

		deliveries = append(deliveries, &delivery)
	}

	return deliveries, nil
}

// Delete deletes a notification delivery record
func (r *notificationDeliveryRepo) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM notification_deliveries WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification delivery: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification delivery not found")
	}

	return nil
}
