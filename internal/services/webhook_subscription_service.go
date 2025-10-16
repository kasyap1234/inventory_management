package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

// WebhookSubscriptionService interface for webhook subscription operations
type WebhookSubscriptionService interface {
	CreateWebhook(ctx context.Context, tenantID uuid.UUID, webhook *models.WebhookSubscription) error
	UpdateWebhook(ctx context.Context, tenantID uuid.UUID, webhook *models.WebhookSubscription) error
	GetWebhook(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.WebhookSubscription, error)
	ListWebhooks(ctx context.Context, tenantID uuid.UUID) ([]*models.WebhookSubscription, error)
	DeleteWebhook(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
	ValidateWebhookURL(url string) error
	TestWebhookConnection(ctx context.Context, webhook *models.WebhookSubscription) error
	DeliverWebhook(ctx context.Context, webhook *models.WebhookSubscription, event *models.NotificationEvent) error
}

// WebhookTestResult represents the result of webhook testing
type WebhookTestResult struct {
	Success      bool   `json:"success"`
	StatusCode   int    `json:"status_code,omitempty"`
	ResponseBody string `json:"response_body,omitempty"`
	Error        string `json:"error,omitempty"`
	Duration     string `json:"duration,omitempty"`
}

// webhookSubscriptionService implements WebhookSubscriptionService
type webhookSubscriptionService struct {
	repository repositories.WebhookSubscriptionRepository
	logger     *common.StructuredLogger
	httpClient *http.Client
}

// NewWebhookSubscriptionService creates a new webhook subscription service
func NewWebhookSubscriptionService(
	repository repositories.WebhookSubscriptionRepository,
	logger *common.StructuredLogger,
) WebhookSubscriptionService {
	return &webhookSubscriptionService{
		repository: repository,
		logger:     logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateWebhook creates a new webhook subscription
func (s *webhookSubscriptionService) CreateWebhook(ctx context.Context, tenantID uuid.UUID, webhook *models.WebhookSubscription) error {
	webhook.TenantID = tenantID
	webhook.CreatedAt = time.Now()
	webhook.UpdatedAt = time.Now()

	// Validate webhook
	if err := s.validateWebhook(webhook); err != nil {
		return common.CreateValidationError("create_webhook", map[string]interface{}{
			"validation": err.Error(),
		})
	}

	// Test webhook connection if active
	if webhook.IsActive {
		if err := s.TestWebhookConnection(ctx, webhook); err != nil {
			s.logger.WarnWithContext(ctx, "Webhook connection test failed during creation", map[string]interface{}{
				"webhook_url": webhook.URL,
				"error":       err.Error(),
			})
		}
	}

	// Create webhook
	if err := s.repository.Create(ctx, webhook); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to create webhook subscription", err, map[string]interface{}{
			"tenant_id":    tenantID,
			"webhook_name": webhook.Name,
			"webhook_url":  webhook.URL,
		})
		return common.CreateDatabaseError("create_webhook", err)
	}

	s.logger.InfoWithContext(ctx, "Webhook subscription created", map[string]interface{}{
		"webhook_id":   webhook.ID,
		"tenant_id":    tenantID,
		"webhook_name": webhook.Name,
		"webhook_url":  webhook.URL,
		"event_types":  webhook.EventTypes,
	})

	// Audit log
	common.AuditCreate(ctx, "webhook_subscription", webhook.ID.String(), map[string]interface{}{
		"name":        webhook.Name,
		"url":         webhook.URL,
		"event_types": webhook.EventTypes,
		"is_active":   webhook.IsActive,
	})

	return nil
}

// UpdateWebhook updates an existing webhook subscription
func (s *webhookSubscriptionService) UpdateWebhook(ctx context.Context, tenantID uuid.UUID, webhook *models.WebhookSubscription) error {
	// Get existing webhook for audit logging
	existing, err := s.repository.GetByID(ctx, tenantID, webhook.ID)
	if err != nil {
		return common.CreateDatabaseError("update_webhook", err)
	}

	webhook.TenantID = tenantID
	webhook.UpdatedAt = time.Now()

	// Validate webhook
	if err := s.validateWebhook(webhook); err != nil {
		return common.CreateValidationError("update_webhook", map[string]interface{}{
			"validation": err.Error(),
		})
	}

	// Test webhook connection if active and URL changed
	if webhook.IsActive && webhook.URL != existing.URL {
		if err := s.TestWebhookConnection(ctx, webhook); err != nil {
			s.logger.WarnWithContext(ctx, "Webhook connection test failed during update", map[string]interface{}{
				"webhook_id":  webhook.ID,
				"webhook_url": webhook.URL,
				"error":       err.Error(),
			})
		}
	}

	// Update webhook
	if err := s.repository.Update(ctx, webhook); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to update webhook subscription", err, map[string]interface{}{
			"webhook_id": webhook.ID,
			"tenant_id":  tenantID,
		})
		return common.CreateDatabaseError("update_webhook", err)
	}

	s.logger.InfoWithContext(ctx, "Webhook subscription updated", map[string]interface{}{
		"webhook_id":   webhook.ID,
		"tenant_id":    tenantID,
		"webhook_name": webhook.Name,
	})

	// Audit log
	oldValues := map[string]interface{}{
		"name":        existing.Name,
		"url":         existing.URL,
		"event_types": existing.EventTypes,
		"is_active":   existing.IsActive,
	}
	newValues := map[string]interface{}{
		"name":        webhook.Name,
		"url":         webhook.URL,
		"event_types": webhook.EventTypes,
		"is_active":   webhook.IsActive,
	}
	common.AuditUpdate(ctx, "webhook_subscription", webhook.ID.String(), oldValues, newValues)

	return nil
}

// GetWebhook retrieves a webhook subscription by ID
func (s *webhookSubscriptionService) GetWebhook(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.WebhookSubscription, error) {
	webhook, err := s.repository.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, common.CreateDatabaseError("get_webhook", err)
	}

	return webhook, nil
}

// ListWebhooks retrieves all webhook subscriptions for a tenant
func (s *webhookSubscriptionService) ListWebhooks(ctx context.Context, tenantID uuid.UUID) ([]*models.WebhookSubscription, error) {
	webhooks, err := s.repository.List(ctx, tenantID)
	if err != nil {
		return nil, common.CreateDatabaseError("list_webhooks", err)
	}

	return webhooks, nil
}

// DeleteWebhook deletes a webhook subscription
func (s *webhookSubscriptionService) DeleteWebhook(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	// Get existing webhook for audit logging
	existing, err := s.repository.GetByID(ctx, tenantID, id)
	if err != nil {
		return common.CreateDatabaseError("delete_webhook", err)
	}

	// Delete webhook
	if err := s.repository.Delete(ctx, tenantID, id); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to delete webhook subscription", err, map[string]interface{}{
			"webhook_id": id,
			"tenant_id":  tenantID,
		})
		return common.CreateDatabaseError("delete_webhook", err)
	}

	s.logger.InfoWithContext(ctx, "Webhook subscription deleted", map[string]interface{}{
		"webhook_id":   id,
		"tenant_id":    tenantID,
		"webhook_name": existing.Name,
	})

	// Audit log
	common.AuditDelete(ctx, "webhook_subscription", id.String(), map[string]interface{}{
		"name":        existing.Name,
		"url":         existing.URL,
		"event_types": existing.EventTypes,
	})

	return nil
}

// ValidateWebhookURL validates a webhook URL
func (s *webhookSubscriptionService) ValidateWebhookURL(webhookURL string) error {
	// Parse URL
	parsedURL, err := url.Parse(webhookURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTP or HTTPS scheme")
	}

	// Prefer HTTPS
	if parsedURL.Scheme == "http" {
		return fmt.Errorf("HTTPS is recommended for webhook URLs")
	}

	// Check host
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	// Prevent localhost and private IPs in production
	if s.isLocalOrPrivateURL(parsedURL.Host) {
		return fmt.Errorf("localhost and private IP addresses are not allowed")
	}

	return nil
}

// TestWebhookConnection tests a webhook connection
func (s *webhookSubscriptionService) TestWebhookConnection(ctx context.Context, webhook *models.WebhookSubscription) error {
	// Create test payload
	testEvent := &models.NotificationEvent{
		Type:      "webhook_test",
		TenantID:  webhook.TenantID,
		Data:      map[string]interface{}{"test": true, "message": "This is a test webhook"},
		Timestamp: time.Now(),
	}

	// Create HTTP request
	payload, err := json.Marshal(testEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal test payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AgroMart-Webhook/1.0")
	req.Header.Set("X-Webhook-Test", "true")

	// Add custom headers
	for key, value := range webhook.Headers {
		if valueStr, ok := value.(string); ok {
			req.Header.Set(key, valueStr)
		}
	}

	// Add signature if secret is provided
	if webhook.Secret != nil && *webhook.Secret != "" {
		signature := s.generateSignature(payload, *webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Set timeout
	client := &http.Client{
		Timeout: time.Duration(webhook.TimeoutSeconds) * time.Second,
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook test request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook test failed with status %d", resp.StatusCode)
	}

	return nil
}

// DeliverWebhook delivers a webhook notification
func (s *webhookSubscriptionService) DeliverWebhook(ctx context.Context, webhook *models.WebhookSubscription, event *models.NotificationEvent) error {
	start := time.Now()

	// Create payload
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AgroMart-Webhook/1.0")
	req.Header.Set("X-Event-Type", event.Type)
	req.Header.Set("X-Tenant-ID", event.TenantID.String())

	// Add custom headers
	for key, value := range webhook.Headers {
		if valueStr, ok := value.(string); ok {
			req.Header.Set(key, valueStr)
		}
	}

	// Add signature if secret is provided
	if webhook.Secret != nil && *webhook.Secret != "" {
		signature := s.generateSignature(payload, *webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Set timeout
	client := &http.Client{
		Timeout: time.Duration(webhook.TimeoutSeconds) * time.Second,
	}

	// Make request
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		// Update failure status
		s.repository.UpdateDeliveryStatus(ctx, webhook.TenantID, webhook.ID, false)
		
		s.logger.ErrorWithContext(ctx, "Webhook delivery failed", err, map[string]interface{}{
			"webhook_id":  webhook.ID,
			"webhook_url": webhook.URL,
			"event_type":  event.Type,
			"duration":    duration.String(),
		})
		
		return fmt.Errorf("webhook delivery failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	
	// Update delivery status
	s.repository.UpdateDeliveryStatus(ctx, webhook.TenantID, webhook.ID, success)

	if success {
		s.logger.InfoWithContext(ctx, "Webhook delivered successfully", map[string]interface{}{
			"webhook_id":  webhook.ID,
			"webhook_url": webhook.URL,
			"event_type":  event.Type,
			"status_code": resp.StatusCode,
			"duration":    duration.String(),
		})
	} else {
		s.logger.WarnWithContext(ctx, "Webhook delivery returned error status", map[string]interface{}{
			"webhook_id":  webhook.ID,
			"webhook_url": webhook.URL,
			"event_type":  event.Type,
			"status_code": resp.StatusCode,
			"duration":    duration.String(),
		})
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// Helper methods

// validateWebhook validates a webhook subscription
func (s *webhookSubscriptionService) validateWebhook(webhook *models.WebhookSubscription) error {
	// Basic validation
	if err := webhook.ValidateWebhook(); err != nil {
		return err
	}

	// Validate URL
	if err := s.ValidateWebhookURL(webhook.URL); err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	// Validate event types
	for _, eventType := range webhook.EventTypes {
		if !s.isValidEventType(eventType) {
			return fmt.Errorf("invalid event type: %s", eventType)
		}
	}

	return nil
}

// isValidEventType validates event type format
func (s *webhookSubscriptionService) isValidEventType(eventType string) bool {
	// Same validation as template service
	if len(eventType) < 3 || len(eventType) > 50 {
		return false
	}
	
	// Check for valid characters
	for _, char := range eventType {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	
	return true
}

// isLocalOrPrivateURL checks if URL points to localhost or private IP
func (s *webhookSubscriptionService) isLocalOrPrivateURL(host string) bool {
	// Remove port if present
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// Check for localhost
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}

	// Check for private IP ranges (simplified check)
	if strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "172.") {
		return true
	}

	return false
}

// generateSignature generates HMAC signature for webhook payload
func (s *webhookSubscriptionService) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}