package repositories

import (
	"context"
	"fmt"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationTemplateRepository interface for notification template operations
type NotificationTemplateRepository interface {
	Create(ctx context.Context, template *models.NotificationTemplate) error
	Update(ctx context.Context, template *models.NotificationTemplate) error
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.NotificationTemplate, error)
	List(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.NotificationTemplate, error)
	ListByType(ctx context.Context, tenantID uuid.UUID, notificationType models.NotificationType) ([]*models.NotificationTemplate, error)
	Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
	GetActiveTemplates(ctx context.Context, tenantID uuid.UUID, eventType string, notificationType models.NotificationType) ([]*models.NotificationTemplate, error)
}

// notificationTemplateRepo implements NotificationTemplateRepository
type notificationTemplateRepo struct {
	db *pgxpool.Pool
}

// NewNotificationTemplateRepo creates a new notification template repository
func NewNotificationTemplateRepo(db *pgxpool.Pool) NotificationTemplateRepository {
	return &notificationTemplateRepo{db: db}
}

// Create creates a new notification template
func (r *notificationTemplateRepo) Create(ctx context.Context, template *models.NotificationTemplate) error {
	// Generate ID if not set
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}

	// Marshal variables to JSON
	variablesJSON, err := template.MarshalVariables()
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	query := `
		INSERT INTO notification_templates (
			id, tenant_id, name, type, event_type, subject, body_template, 
			variables, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
		)`

	_, err = r.db.Exec(ctx, query,
		template.ID,
		template.TenantID,
		template.Name,
		string(template.Type),
		template.EventType,
		template.Subject,
		template.BodyTemplate,
		variablesJSON,
		template.IsActive,
	)

	if err != nil {
		if err.Error() == "SQLSTATE 23505" || fmt.Sprintf("%v", err) == "duplicate key value violates unique constraint" {
			return fmt.Errorf("notification template with name '%s' and type '%s' already exists", template.Name, template.Type)
		}
		return fmt.Errorf("failed to create notification template: %w", err)
	}

	return nil
}

// Update updates an existing notification template
func (r *notificationTemplateRepo) Update(ctx context.Context, template *models.NotificationTemplate) error {
	// Marshal variables to JSON
	variablesJSON, err := template.MarshalVariables()
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	query := `
		UPDATE notification_templates SET
			name = $3,
			type = $4,
			event_type = $5,
			subject = $6,
			body_template = $7,
			variables = $8,
			is_active = $9,
			updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query,
		template.TenantID,
		template.ID,
		template.Name,
		string(template.Type),
		template.EventType,
		template.Subject,
		template.BodyTemplate,
		variablesJSON,
		template.IsActive,
	)

	if err != nil {
		if err.Error() == "SQLSTATE 23505" || fmt.Sprintf("%v", err) == "duplicate key value violates unique constraint" {
			return fmt.Errorf("notification template with name '%s' and type '%s' already exists", template.Name, template.Type)
		}
		return fmt.Errorf("failed to update notification template: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification template not found")
	}

	return nil
}

// GetByID retrieves a notification template by ID
func (r *notificationTemplateRepo) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.NotificationTemplate, error) {
	query := `
		SELECT id, tenant_id, name, type, event_type, subject, body_template,
			   variables, is_active, created_at, updated_at
		FROM notification_templates
		WHERE tenant_id = $1 AND id = $2`

	var template models.NotificationTemplate
	var variablesJSON []byte

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&template.ID,
		&template.TenantID,
		&template.Name,
		&template.Type,
		&template.EventType,
		&template.Subject,
		&template.BodyTemplate,
		&variablesJSON,
		&template.IsActive,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("notification template not found")
		}
		return nil, fmt.Errorf("failed to get notification template: %w", err)
	}

	// Unmarshal variables
	if err := template.UnmarshalVariables(variablesJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}

	return &template, nil
}

// List retrieves notification templates, optionally filtered by event type
func (r *notificationTemplateRepo) List(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.NotificationTemplate, error) {
	query := `
		SELECT id, tenant_id, name, type, event_type, subject, body_template,
			   variables, is_active, created_at, updated_at
		FROM notification_templates
		WHERE tenant_id = $1`

	args := []interface{}{tenantID}

	if eventType != "" {
		query += " AND event_type = $2"
		args = append(args, eventType)
	}

	query += " ORDER BY name ASC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notification templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.NotificationTemplate

	for rows.Next() {
		var template models.NotificationTemplate
		var variablesJSON []byte

		err := rows.Scan(
			&template.ID,
			&template.TenantID,
			&template.Name,
			&template.Type,
			&template.EventType,
			&template.Subject,
			&template.BodyTemplate,
			&variablesJSON,
			&template.IsActive,
			&template.CreatedAt,
			&template.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan notification template: %w", err)
		}

		// Unmarshal variables
		if err := template.UnmarshalVariables(variablesJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}

		templates = append(templates, &template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notification templates: %w", err)
	}

	return templates, nil
}

// ListByType retrieves notification templates by type
func (r *notificationTemplateRepo) ListByType(ctx context.Context, tenantID uuid.UUID, notificationType models.NotificationType) ([]*models.NotificationTemplate, error) {
	query := `
		SELECT id, tenant_id, name, type, event_type, subject, body_template,
			   variables, is_active, created_at, updated_at
		FROM notification_templates
		WHERE tenant_id = $1 AND type = $2
		ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query, tenantID, string(notificationType))
	if err != nil {
		return nil, fmt.Errorf("failed to list notification templates by type: %w", err)
	}
	defer rows.Close()

	var templates []*models.NotificationTemplate

	for rows.Next() {
		var template models.NotificationTemplate
		var variablesJSON []byte

		err := rows.Scan(
			&template.ID,
			&template.TenantID,
			&template.Name,
			&template.Type,
			&template.EventType,
			&template.Subject,
			&template.BodyTemplate,
			&variablesJSON,
			&template.IsActive,
			&template.CreatedAt,
			&template.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan notification template: %w", err)
		}

		// Unmarshal variables
		if err := template.UnmarshalVariables(variablesJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}

		templates = append(templates, &template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notification templates: %w", err)
	}

	return templates, nil
}

// Delete deletes a notification template
func (r *notificationTemplateRepo) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM notification_templates WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification template: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification template not found")
	}

	return nil
}

// GetActiveTemplates retrieves active templates for a specific event type and notification type
func (r *notificationTemplateRepo) GetActiveTemplates(ctx context.Context, tenantID uuid.UUID, eventType string, notificationType models.NotificationType) ([]*models.NotificationTemplate, error) {
	query := `
		SELECT id, tenant_id, name, type, event_type, subject, body_template,
			   variables, is_active, created_at, updated_at
		FROM notification_templates
		WHERE tenant_id = $1 AND event_type = $2 AND type = $3 AND is_active = true
		ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query, tenantID, eventType, string(notificationType))
	if err != nil {
		return nil, fmt.Errorf("failed to get active notification templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.NotificationTemplate

	for rows.Next() {
		var template models.NotificationTemplate
		var variablesJSON []byte

		err := rows.Scan(
			&template.ID,
			&template.TenantID,
			&template.Name,
			&template.Type,
			&template.EventType,
			&template.Subject,
			&template.BodyTemplate,
			&variablesJSON,
			&template.IsActive,
			&template.CreatedAt,
			&template.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan notification template: %w", err)
		}

		// Unmarshal variables
		if err := template.UnmarshalVariables(variablesJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}

		templates = append(templates, &template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active notification templates: %w", err)
	}

	return templates, nil
}
