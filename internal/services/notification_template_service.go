package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

// NotificationTemplateService interface for notification template operations
type NotificationTemplateService interface {
	CreateTemplate(ctx context.Context, template *models.NotificationTemplate) error
	UpdateTemplate(ctx context.Context, template *models.NotificationTemplate) error
	GetTemplate(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.NotificationTemplate, error)
	ListTemplates(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.NotificationTemplate, error)
	DeleteTemplate(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
	ValidateTemplate(template *models.NotificationTemplate) error
	RenderTemplate(ctx context.Context, template *models.NotificationTemplate, data map[string]interface{}) (string, string, error)
	RenderTemplateWithValidation(template *models.NotificationTemplate, data map[string]interface{}) (string, string, error)
	GetActiveTemplatesForEvent(ctx context.Context, tenantID uuid.UUID, eventType string, notificationType models.NotificationType) ([]*models.NotificationTemplate, error)
	TestTemplate(ctx context.Context, tenantID uuid.UUID, templateID uuid.UUID, testData map[string]interface{}) (*TemplateTestResult, error)
}

// TemplateTestResult represents the result of template testing
type TemplateTestResult struct {
	Success      bool                   `json:"success"`
	RenderedBody string                 `json:"rendered_body,omitempty"`
	RenderedSubject string              `json:"rendered_subject,omitempty"`
	Errors       []string               `json:"errors,omitempty"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
}

// notificationTemplateService implements NotificationTemplateService
type notificationTemplateService struct {
	repository repositories.NotificationTemplateRepository
	logger     *common.StructuredLogger
}

// NewNotificationTemplateService creates a new notification template service
func NewNotificationTemplateService(
	repository repositories.NotificationTemplateRepository,
) NotificationTemplateService {
	return &notificationTemplateService{
		repository: repository,
		logger:     common.NewStructuredLogger(),
	}
}

// CreateTemplate creates a new notification template
func (s *notificationTemplateService) CreateTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	// Set timestamps
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	// Validate template
	if err := s.ValidateTemplate(template); err != nil {
		return common.CreateValidationError("create_template", map[string]interface{}{
			"validation": err.Error(),
		})
	}

	// Create template
	if err := s.repository.Create(ctx, template); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to create notification template", err, map[string]interface{}{
			"tenant_id":    template.TenantID,
			"template_name": template.Name,
			"event_type":   template.EventType,
		})
		return common.CreateDatabaseError("create_template", err)
	}

	// Log successful creation
	s.logger.InfoWithContext(ctx, "Notification template created", map[string]interface{}{
		"template_id":   template.ID,
		"tenant_id":     template.TenantID,
		"template_name": template.Name,
		"event_type":    template.EventType,
		"type":          template.Type,
	})

	// Audit log
	common.AuditCreate(ctx, "notification_template", template.ID.String(), map[string]interface{}{
		"name":       template.Name,
		"type":       template.Type,
		"event_type": template.EventType,
		"is_active":  template.IsActive,
	})

	return nil
}

// UpdateTemplate updates an existing notification template
func (s *notificationTemplateService) UpdateTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	// Get existing template for audit logging
	existing, err := s.repository.GetByID(ctx, template.TenantID, template.ID)
	if err != nil {
		return common.CreateDatabaseError("update_template", err)
	}

	// Set update timestamp
	template.UpdatedAt = time.Now()

	// Validate template
	if err := s.ValidateTemplate(template); err != nil {
		return common.CreateValidationError("update_template", map[string]interface{}{
			"validation": err.Error(),
		})
	}

	// Update template
	if err := s.repository.Update(ctx, template); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to update notification template", err, map[string]interface{}{
			"template_id": template.ID,
			"tenant_id":   template.TenantID,
		})
		return common.CreateDatabaseError("update_template", err)
	}

	// Log successful update
	s.logger.InfoWithContext(ctx, "Notification template updated", map[string]interface{}{
		"template_id":   template.ID,
		"tenant_id":     template.TenantID,
		"template_name": template.Name,
	})

	// Audit log
	oldValues := map[string]interface{}{
		"name":          existing.Name,
		"type":          existing.Type,
		"event_type":    existing.EventType,
		"subject":       existing.Subject,
		"body_template": existing.BodyTemplate,
		"is_active":     existing.IsActive,
	}
	newValues := map[string]interface{}{
		"name":          template.Name,
		"type":          template.Type,
		"event_type":    template.EventType,
		"subject":       template.Subject,
		"body_template": template.BodyTemplate,
		"is_active":     template.IsActive,
	}
	common.AuditUpdate(ctx, "notification_template", template.ID.String(), oldValues, newValues)

	return nil
}

// GetTemplate retrieves a notification template by ID
func (s *notificationTemplateService) GetTemplate(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.NotificationTemplate, error) {
	template, err := s.repository.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, common.CreateDatabaseError("get_template", err)
	}

	return template, nil
}

// ListTemplates retrieves notification templates
func (s *notificationTemplateService) ListTemplates(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.NotificationTemplate, error) {
	templates, err := s.repository.List(ctx, tenantID, eventType)
	if err != nil {
		return nil, common.CreateDatabaseError("list_templates", err)
	}

	return templates, nil
}

// DeleteTemplate deletes a notification template
func (s *notificationTemplateService) DeleteTemplate(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	// Get existing template for audit logging
	existing, err := s.repository.GetByID(ctx, tenantID, id)
	if err != nil {
		return common.CreateDatabaseError("delete_template", err)
	}

	// Delete template
	if err := s.repository.Delete(ctx, tenantID, id); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to delete notification template", err, map[string]interface{}{
			"template_id": id,
			"tenant_id":   tenantID,
		})
		return common.CreateDatabaseError("delete_template", err)
	}

	// Log successful deletion
	s.logger.InfoWithContext(ctx, "Notification template deleted", map[string]interface{}{
		"template_id":   id,
		"tenant_id":     tenantID,
		"template_name": existing.Name,
	})

	// Audit log
	common.AuditDelete(ctx, "notification_template", id.String(), map[string]interface{}{
		"name":       existing.Name,
		"type":       existing.Type,
		"event_type": existing.EventType,
	})

	return nil
}

// ValidateTemplate validates a notification template
func (s *notificationTemplateService) ValidateTemplate(template *models.NotificationTemplate) error {
	// Basic validation
	if err := template.ValidateTemplate(); err != nil {
		return err
	}

	// Validate template syntax
	if err := s.validateTemplateSyntax(template.BodyTemplate); err != nil {
		return fmt.Errorf("invalid body template syntax: %w", err)
	}

	// Validate subject template syntax for email templates
	if template.Type == models.NotificationTypeEmail && template.Subject != nil {
		if err := s.validateTemplateSyntax(*template.Subject); err != nil {
			return fmt.Errorf("invalid subject template syntax: %w", err)
		}
	}

	// Validate event type format
	if !s.isValidEventType(template.EventType) {
		return fmt.Errorf("invalid event type format: %s", template.EventType)
	}

	// Validate variables
	if err := s.validateTemplateVariables(template.Variables); err != nil {
		return fmt.Errorf("invalid template variables: %w", err)
	}

	return nil
}

// RenderTemplate renders a template with the given data
func (s *notificationTemplateService) RenderTemplate(ctx context.Context, template *models.NotificationTemplate, data map[string]interface{}) (string, string, error) {
	return s.RenderTemplateWithValidation(template, data)
}

// RenderTemplateWithValidation renders a template with validation
func (s *notificationTemplateService) RenderTemplateWithValidation(template *models.NotificationTemplate, data map[string]interface{}) (string, string, error) {
	// Render body template
	renderedBody, err := s.renderTemplateString(template.BodyTemplate, data)
	if err != nil {
		return "", "", fmt.Errorf("failed to render body template: %w", err)
	}

	// Render subject template (for email templates)
	var renderedSubject string
	if template.Type == models.NotificationTypeEmail && template.Subject != nil {
		renderedSubject, err = s.renderTemplateString(*template.Subject, data)
		if err != nil {
			return "", "", fmt.Errorf("failed to render subject template: %w", err)
		}
	}

	return renderedBody, renderedSubject, nil
}

// GetActiveTemplatesForEvent retrieves active templates for a specific event
func (s *notificationTemplateService) GetActiveTemplatesForEvent(ctx context.Context, tenantID uuid.UUID, eventType string, notificationType models.NotificationType) ([]*models.NotificationTemplate, error) {
	templates, err := s.repository.GetActiveTemplates(ctx, tenantID, eventType, notificationType)
	if err != nil {
		return nil, common.CreateDatabaseError("get_active_templates", err)
	}

	return templates, nil
}

// TestTemplate tests a template with sample data
func (s *notificationTemplateService) TestTemplate(ctx context.Context, tenantID uuid.UUID, templateID uuid.UUID, testData map[string]interface{}) (*TemplateTestResult, error) {
	// Get template
	template, err := s.repository.GetByID(ctx, tenantID, templateID)
	if err != nil {
		return nil, common.CreateDatabaseError("test_template", err)
	}

	result := &TemplateTestResult{
		Variables: template.Variables,
	}

	// Try to render the template
	renderedBody, renderedSubject, err := s.RenderTemplateWithValidation(template, testData)
	if err != nil {
		result.Success = false
		result.Errors = []string{err.Error()}
		return result, nil
	}

	result.Success = true
	result.RenderedBody = renderedBody
	result.RenderedSubject = renderedSubject

	return result, nil
}

// Helper methods

// validateTemplateSyntax validates Go template syntax
func (s *notificationTemplateService) validateTemplateSyntax(templateStr string) error {
	_, err := template.New("test").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("template syntax error: %w", err)
	}
	return nil
}

// isValidEventType validates event type format
func (s *notificationTemplateService) isValidEventType(eventType string) bool {
	// Event types should follow a pattern like "order_created", "low_stock", etc.
	pattern := `^[a-z][a-z0-9_]*[a-z0-9]$`
	matched, err := regexp.MatchString(pattern, eventType)
	if err != nil {
		return false
	}
	return matched && len(eventType) >= 3 && len(eventType) <= 50
}

// validateTemplateVariables validates template variables
func (s *notificationTemplateService) validateTemplateVariables(variables map[string]interface{}) error {
	if variables == nil {
		return nil
	}

	for key, value := range variables {
		// Validate variable name
		if !s.isValidVariableName(key) {
			return fmt.Errorf("invalid variable name: %s", key)
		}

		// Validate variable value structure
		if err := s.validateVariableValue(value); err != nil {
			return fmt.Errorf("invalid variable '%s': %w", key, err)
		}
	}

	return nil
}

// isValidVariableName validates variable name format
func (s *notificationTemplateService) isValidVariableName(name string) bool {
	pattern := `^[a-zA-Z][a-zA-Z0-9_]*$`
	matched, err := regexp.MatchString(pattern, name)
	if err != nil {
		return false
	}
	return matched && len(name) >= 1 && len(name) <= 50
}

// validateVariableValue validates variable value structure
func (s *notificationTemplateService) validateVariableValue(value interface{}) error {
	switch v := value.(type) {
	case string:
		if len(v) > 1000 {
			return fmt.Errorf("string value too long (max 1000 characters)")
		}
	case map[string]interface{}:
		// Validate nested structure
		for key, val := range v {
			if !s.isValidVariableName(key) {
				return fmt.Errorf("invalid nested variable name: %s", key)
			}
			if err := s.validateVariableValue(val); err != nil {
				return fmt.Errorf("invalid nested variable '%s': %w", key, err)
			}
		}
	case []interface{}:
		// Validate array elements
		for i, val := range v {
			if err := s.validateVariableValue(val); err != nil {
				return fmt.Errorf("invalid array element at index %d: %w", i, err)
			}
		}
	case float64, int, bool:
		// These are valid types
	default:
		return fmt.Errorf("unsupported variable type: %T", value)
	}

	return nil
}

// renderTemplateString renders a template string with data
func (s *notificationTemplateService) renderTemplateString(templateStr string, data map[string]interface{}) (string, error) {
	// Create template with helper functions
	tmpl := template.New("notification").Funcs(template.FuncMap{
		"upper":     strings.ToUpper,
		"lower":     strings.ToLower,
		"title":     strings.Title,
		"trim":      strings.TrimSpace,
		"formatDate": func(t time.Time, layout string) string {
			return t.Format(layout)
		},
		"formatCurrency": func(amount float64, currency string) string {
			return fmt.Sprintf("%.2f %s", amount, currency)
		},
		"default": func(defaultValue, value interface{}) interface{} {
			if value == nil || value == "" {
				return defaultValue
			}
			return value
		},
	})

	// Parse template
	tmpl, err := tmpl.Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}