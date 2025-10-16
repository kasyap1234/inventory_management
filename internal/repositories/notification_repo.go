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

// NotificationRepository handles enhanced notifications persistence
type NotificationRepository interface {
	Create(ctx context.Context, notification *models.EnhancedNotification) error
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.EnhancedNotification, error)
	Update(ctx context.Context, notification *models.EnhancedNotification) error
	Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filters NotificationFilters) ([]*models.EnhancedNotification, error)
	MarkAsRead(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error
	DeleteOld(ctx context.Context, tenantID uuid.UUID, olderThan time.Time) error
}

// NotificationFilters represents filters for querying notifications
type NotificationFilters struct {
	UserID             *uuid.UUID `json:"user_id,omitempty"`
	NotificationType   *string    `json:"notification_type,omitempty"`
	EventType          *string    `json:"event_type,omitempty"`
	Status             *string    `json:"status,omitempty"`
	Priority           *string    `json:"priority,omitempty"`
	StartDate          *time.Time `json:"start_date,omitempty"`
	EndDate            *time.Time `json:"end_date,omitempty"`
	Limit              int        `json:"limit,omitempty"`
	Offset             int        `json:"offset,omitempty"`
}

// notificationRepo implements NotificationRepository
type notificationRepo struct {
	db *pgxpool.Pool
}

// NewNotificationRepo creates a new notification repository
func NewNotificationRepo(db *pgxpool.Pool) NotificationRepository {
	return &notificationRepo{db: db}
}

// Create creates a new notification
func (r *notificationRepo) Create(ctx context.Context, notification *models.EnhancedNotification) error {
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}

	now := time.Now()
	notification.CreatedAt = now
	notification.UpdatedAt = now

	eventDataJSON, err := json.Marshal(notification.EventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	query := `
		INSERT INTO notifications (
			id, tenant_id, user_id, title, message, notification_type,
			priority, status, event_type, event_data, template_id,
			read_at, expires_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)`

	_, err = r.db.Exec(ctx, query,
		notification.ID, notification.TenantID, notification.UserID,
		notification.Title, notification.Message, notification.NotificationType,
		notification.Priority, notification.Status, notification.EventType,
		eventDataJSON, notification.TemplateID, notification.ReadAt,
		notification.ExpiresAt, notification.CreatedAt, notification.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// GetByID retrieves a notification by ID
func (r *notificationRepo) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.EnhancedNotification, error) {
	query := `
		SELECT id, tenant_id, user_id, title, message, notification_type,
			   priority, status, event_type, event_data, template_id,
			   read_at, expires_at, created_at, updated_at
		FROM notifications
		WHERE tenant_id = $1 AND id = $2`

	var notification models.EnhancedNotification
	var eventDataJSON []byte

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&notification.ID, &notification.TenantID, &notification.UserID,
		&notification.Title, &notification.Message, &notification.NotificationType,
		&notification.Priority, &notification.Status, &notification.EventType,
		&eventDataJSON, &notification.TemplateID, &notification.ReadAt,
		&notification.ExpiresAt, &notification.CreatedAt, &notification.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	if len(eventDataJSON) > 0 {
		if err := json.Unmarshal(eventDataJSON, &notification.EventData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}
	}

	return &notification, nil
}

// Update updates an existing notification
func (r *notificationRepo) Update(ctx context.Context, notification *models.EnhancedNotification) error {
	notification.UpdatedAt = time.Now()

	eventDataJSON, err := json.Marshal(notification.EventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	query := `
		UPDATE notifications SET
			title = $3, message = $4, notification_type = $5,
			priority = $6, status = $7, event_type = $8, event_data = $9,
			template_id = $10, read_at = $11, expires_at = $12, updated_at = $13
		WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query,
		notification.TenantID, notification.ID, notification.Title,
		notification.Message, notification.NotificationType, notification.Priority,
		notification.Status, notification.EventType, eventDataJSON,
		notification.TemplateID, notification.ReadAt, notification.ExpiresAt,
		notification.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// Delete deletes a notification
func (r *notificationRepo) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// List retrieves notifications with filters
func (r *notificationRepo) List(ctx context.Context, tenantID uuid.UUID, filters NotificationFilters) ([]*models.EnhancedNotification, error) {
	query := `
		SELECT id, tenant_id, user_id, title, message, notification_type,
			   priority, status, event_type, event_data, template_id,
			   read_at, expires_at, created_at, updated_at
		FROM notifications
		WHERE tenant_id = $1`

	args := []interface{}{tenantID}
	argIndex := 2

	// Add filters
	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filters.UserID)
		argIndex++
	}

	if filters.NotificationType != nil {
		query += fmt.Sprintf(" AND notification_type = $%d", argIndex)
		args = append(args, *filters.NotificationType)
		argIndex++
	}

	if filters.EventType != nil {
		query += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, *filters.EventType)
		argIndex++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.Priority != nil {
		query += fmt.Sprintf(" AND priority = $%d", argIndex)
		args = append(args, *filters.Priority)
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
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
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
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*models.EnhancedNotification

	for rows.Next() {
		var notification models.EnhancedNotification
		var eventDataJSON []byte

		err := rows.Scan(
			&notification.ID, &notification.TenantID, &notification.UserID,
			&notification.Title, &notification.Message, &notification.NotificationType,
			&notification.Priority, &notification.Status, &notification.EventType,
			&eventDataJSON, &notification.TemplateID, &notification.ReadAt,
			&notification.ExpiresAt, &notification.CreatedAt, &notification.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if len(eventDataJSON) > 0 {
			if err := json.Unmarshal(eventDataJSON, &notification.EventData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
			}
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (r *notificationRepo) MarkAsRead(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	query := `
		UPDATE notifications SET
			status = 'read',
			read_at = NOW(),
			updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *notificationRepo) MarkAllAsRead(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE notifications SET
			status = 'read',
			read_at = NOW(),
			updated_at = NOW()
		WHERE tenant_id = $1 AND user_id = $2 AND status != 'read'`

	_, err := r.db.Exec(ctx, query, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// DeleteOld deletes notifications older than the specified date
func (r *notificationRepo) DeleteOld(ctx context.Context, tenantID uuid.UUID, olderThan time.Time) error {
	query := `DELETE FROM notifications WHERE tenant_id = $1 AND created_at < $2`

	_, err := r.db.Exec(ctx, query, tenantID, olderThan)
	if err != nil {
		return fmt.Errorf("failed to delete old notifications: %w", err)
	}

	return nil
}
