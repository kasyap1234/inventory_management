package repositories

import (
	"context"
	"fmt"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationConfigRepository handles user notification preferences
type NotificationConfigRepository interface {
	Create(ctx context.Context, config *models.NotificationConfig) error
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.NotificationConfig, error)
	GetByUserAndEventType(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, eventType string) (*models.NotificationConfig, error)
	Update(ctx context.Context, config *models.NotificationConfig) error
	Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, limit, offset int) ([]*models.NotificationConfig, error)
	ListByUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) ([]*models.NotificationConfig, error)
}

// notificationConfigRepo implements NotificationConfigRepository
type notificationConfigRepo struct {
	db *pgxpool.Pool
}

// NewNotificationConfigRepo creates a new notification config repository
func NewNotificationConfigRepo(db *pgxpool.Pool) NotificationConfigRepository {
	return &notificationConfigRepo{db: db}
}

// Create creates a new notification config
func (r *notificationConfigRepo) Create(ctx context.Context, config *models.NotificationConfig) error {
	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}

	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	query := `
		INSERT INTO notification_configs (
			id, tenant_id, user_id, event_type, channels, is_enabled, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)`

	_, err := r.db.Exec(ctx, query,
		config.ID, config.TenantID, config.UserID, config.EventType,
		config.Channels, config.IsEnabled, config.CreatedAt, config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification config: %w", err)
	}

	return nil
}

// GetByID retrieves a notification config by ID
func (r *notificationConfigRepo) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.NotificationConfig, error) {
	query := `
		SELECT id, tenant_id, user_id, event_type, channels, is_enabled, created_at, updated_at
		FROM notification_configs
		WHERE tenant_id = $1 AND id = $2`

	var config models.NotificationConfig

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&config.ID, &config.TenantID, &config.UserID, &config.EventType,
		&config.Channels, &config.IsEnabled, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("notification config not found")
		}
		return nil, fmt.Errorf("failed to get notification config: %w", err)
	}

	return &config, nil
}

// GetByUserAndEventType retrieves a notification config for a specific user and event type
func (r *notificationConfigRepo) GetByUserAndEventType(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, eventType string) (*models.NotificationConfig, error) {
	query := `
		SELECT id, tenant_id, user_id, event_type, channels, is_enabled, created_at, updated_at
		FROM notification_configs
		WHERE tenant_id = $1 AND user_id = $2 AND event_type = $3`

	var config models.NotificationConfig

	err := r.db.QueryRow(ctx, query, tenantID, userID, eventType).Scan(
		&config.ID, &config.TenantID, &config.UserID, &config.EventType,
		&config.Channels, &config.IsEnabled, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			// Return default config if not found
			return &models.NotificationConfig{
				ID:        uuid.New(),
				TenantID:  tenantID,
				UserID:    userID,
				EventType: eventType,
				Channels:  []string{"email"}, // Default to email
				IsEnabled: true,
			}, nil
		}
		return nil, fmt.Errorf("failed to get notification config: %w", err)
	}

	return &config, nil
}

// Update updates an existing notification config
func (r *notificationConfigRepo) Update(ctx context.Context, config *models.NotificationConfig) error {
	config.UpdatedAt = time.Now()

	query := `
		UPDATE notification_configs SET
			event_type = $3, channels = $4, is_enabled = $5, updated_at = $6
		WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query,
		config.TenantID, config.ID, config.EventType,
		config.Channels, config.IsEnabled, config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification config: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification config not found")
	}

	return nil
}

// Delete deletes a notification config
func (r *notificationConfigRepo) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM notification_configs WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification config: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification config not found")
	}

	return nil
}

// List retrieves notification configs with pagination
func (r *notificationConfigRepo) List(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, limit, offset int) ([]*models.NotificationConfig, error) {
	query := `
		SELECT id, tenant_id, user_id, event_type, channels, is_enabled, created_at, updated_at
		FROM notification_configs
		WHERE tenant_id = $1`

	args := []interface{}{tenantID}
	argIndex := 2

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	query += fmt.Sprintf(" LIMIT $%d", argIndex)
	args = append(args, limit)
	argIndex++

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notification configs: %w", err)
	}
	defer rows.Close()

	var configs []*models.NotificationConfig

	for rows.Next() {
		var config models.NotificationConfig

		err := rows.Scan(
			&config.ID, &config.TenantID, &config.UserID, &config.EventType,
			&config.Channels, &config.IsEnabled, &config.CreatedAt, &config.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan notification config: %w", err)
		}

		configs = append(configs, &config)
	}

	return configs, nil
}

// ListByUser retrieves all notification configs for a user
func (r *notificationConfigRepo) ListByUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) ([]*models.NotificationConfig, error) {
	query := `
		SELECT id, tenant_id, user_id, event_type, channels, is_enabled, created_at, updated_at
		FROM notification_configs
		WHERE tenant_id = $1 AND user_id = $2
		ORDER BY event_type ASC`

	rows, err := r.db.Query(ctx, query, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user notification configs: %w", err)
	}
	defer rows.Close()

	var configs []*models.NotificationConfig

	for rows.Next() {
		var config models.NotificationConfig

		err := rows.Scan(
			&config.ID, &config.TenantID, &config.UserID, &config.EventType,
			&config.Channels, &config.IsEnabled, &config.CreatedAt, &config.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan notification config: %w", err)
		}

		configs = append(configs, &config)
	}

	return configs, nil
}
