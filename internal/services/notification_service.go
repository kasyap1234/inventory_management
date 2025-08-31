package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// NotificationService handles all notification-related operations
type NotificationService interface {
	// Core notification methods
	SendNotification(ctx context.Context, tenantID uuid.UUID, notification *models.Notification) error
	SendEmail(ctx context.Context, tenantID uuid.UUID, recipient, subject, body string) error
	SendSMS(ctx context.Context, tenantID uuid.UUID, recipient, message string) error
	SendWebhook(ctx context.Context, tenantID uuid.UUID, webhook *models.WebhookSubscription, payload map[string]interface{}) error

	// Template management
	CreateTemplate(ctx context.Context, tenantID uuid.UUID, template *models.NotificationTemplate) error
	UpdateTemplate(ctx context.Context, tenantID uuid.UUID, template *models.NotificationTemplate) error
	DeleteTemplate(ctx context.Context, tenantID uuid.UUID, templateID string) error
	GetTemplate(ctx context.Context, tenantID uuid.UUID, templateID string) (*models.NotificationTemplate, error)
	ListTemplates(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.NotificationTemplate, error)

	// Configuration management
	UpdateNotificationConfig(ctx context.Context, tenantID uuid.UUID, config *models.NotificationConfig) error
	GetNotificationConfig(ctx context.Context, tenantID uuid.UUID, notificationType models.NotificationType) (*models.NotificationConfig, error)

	// Alert management
	CreateAlert(ctx context.Context, tenantID uuid.UUID, alert *models.Alert) error
	UpdateAlertStatus(ctx context.Context, tenantID uuid.UUID, alertID string, status string) error
	ProcessAlert(ctx context.Context, tenantID uuid.UUID, alertID string) error
	CheckAndTriggerAlerts(ctx context.Context, tenantID uuid.UUID) error

	// Webhook subscription management
	CreateWebhookSubscription(ctx context.Context, tenantID uuid.UUID, subscription *models.WebhookSubscription) error
	UpdateWebhookSubscription(ctx context.Context, tenantID uuid.UUID, subscription *models.WebhookSubscription) error
	DeleteWebhookSubscription(ctx context.Context, tenantID uuid.UUID, subscriptionID string) error
	ListWebhookSubscriptions(ctx context.Context, tenantID uuid.UUID) ([]*models.WebhookSubscription, error)

	// Alert configuration
	UpdateAlertConfig(ctx context.Context, tenantID uuid.UUID, config *models.AlertConfig) error
	GetAlertConfig(ctx context.Context, tenantID uuid.UUID, alertType models.AlertType) (*models.AlertConfig, error)

	// Utility methods
	RenderTemplate(template *models.NotificationTemplate, data map[string]interface{}) (string, error)
	RetryFailedNotifications(ctx context.Context) error
}

type notificationService struct {
	redisClient *redis.Client
	templates   map[string]*template.Template // Cached templates
	httpClient  *http.Client
}

// NewNotificationService creates a new notification service
func NewNotificationService(redisAddr, redisPassword string, redisDB int) NotificationService {
	// Create Redis client for this service
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &notificationService{
		redisClient: redisClient,
		templates:   make(map[string]*template.Template),
		httpClient:  httpClient,
	}
}

// SendNotification sends a notification via the configured channel
func (s *notificationService) SendNotification(ctx context.Context, tenantID uuid.UUID, notification *models.Notification) error {
	switch notification.Type {
	case models.NotificationTypeEmail:
		return s.SendEmail(ctx, tenantID, notification.Recipient, *notification.Subject, notification.Body)
	case models.NotificationTypeSMS:
		return s.SendSMS(ctx, tenantID, notification.Recipient, notification.Body)
	case models.NotificationTypeWebhook:
		// For webhook, recipient is treated as webhook subscription ID
		subscription, err := s.getWebhookSubscription(ctx, tenantID, notification.Recipient)
		if err != nil {
			return fmt.Errorf("failed to get webhook subscription: %v", err)
		}

		payload := map[string]interface{}{
			"type":      notification.EventType,
			"event_id":  notification.EventID,
			"subject":   notification.Subject,
			"body":      notification.Body,
			"timestamp": time.Now(),
		}

		return s.SendWebhook(ctx, tenantID, subscription, payload)
	default:
		return fmt.Errorf("unsupported notification type: %s", notification.Type)
	}
}

// SendEmail sends an email notification (placeholder implementation)
func (s *notificationService) SendEmail(ctx context.Context, tenantID uuid.UUID, recipient, subject, body string) error {
	// TODO: Integration with email service (SendGrid, SES, etc.)
	// Placeholder implementation - log the email that would be sent

	log.Printf("[EMAIL] Tenant=%s, To=%s, Subject=%s, Body=%s", tenantID.String(), recipient, subject, body)

	// In production, integrate with actual email provider
	// Example: SendGrid API call would go here

	return nil // Placeholder - no actual sending
}

// SendSMS sends an SMS notification (placeholder implementation)
func (s *notificationService) SendSMS(ctx context.Context, tenantID uuid.UUID, recipient, message string) error {
	// TODO: Integration with SMS service (Twilio, AWS SNS, etc.)
	// Placeholder implementation - log the SMS that would be sent

	log.Printf("[SMS] Tenant=%s, To=%s, Message=%s", tenantID.String(), recipient, message)

	// In production, integrate with actual SMS provider
	// Example: Twilio API call would go here

	return nil // Placeholder - no actual sending
}

// SendWebhook sends a webhook notification
func (s *notificationService) SendWebhook(ctx context.Context, tenantID uuid.UUID, webhook *models.WebhookSubscription, payload map[string]interface{}) error {
	if !webhook.IsActive {
		return nil // Skip inactive webhooks
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", "signature-placeholder") // TODO: Implement HMAC signature
	req.Header.Set("X-Tenant-ID", tenantID.String())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	log.Printf("[WEBHOOK] Tenant=%s, URL=%s, Payload=%s", tenantID.String(), webhook.URL, string(jsonPayload))

	// Update last used timestamp
	webhook.LastUsedAt = &time.Time{}
	*webhook.LastUsedAt = time.Now()
	s.updateWebhookSubscription(ctx, tenantID, webhook)

	return nil
}

// Template management methods
func (s *notificationService) CreateTemplate(ctx context.Context, tenantID uuid.UUID, template *models.NotificationTemplate) error {
	template.ID = uuid.NewString()
	template.TenantID = tenantID.String()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	// Cache the template for faster rendering
	s.cacheTemplate(template)

	// Store in Redis
	cacheKey := fmt.Sprintf("notification_template:%s:%s", tenantID.String(), template.ID)
	data, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %v", err)
	}

	err = s.redisClient.Set(ctx, cacheKey, data, time.Hour).Err()
	if err != nil {
		log.Printf("Failed to cache template: %v", err)
	}

	return nil
}

func (s *notificationService) UpdateTemplate(ctx context.Context, tenantID uuid.UUID, template *models.NotificationTemplate) error {
	template.UpdatedAt = time.Now()
	s.cacheTemplate(template)

	cacheKey := fmt.Sprintf("notification_template:%s:%s", tenantID.String(), template.ID)
	data, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %v", err)
	}

	return s.redisClient.Set(ctx, cacheKey, data, time.Hour).Err()
}

func (s *notificationService) DeleteTemplate(ctx context.Context, tenantID uuid.UUID, templateID string) error {
	cacheKey := fmt.Sprintf("notification_template:%s:%s", tenantID.String(), templateID)
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("Failed to delete cached template: %v", err)
	}

	delete(s.templates, fmt.Sprintf("%s:%s", tenantID.String(), templateID))
	return nil
}

func (s *notificationService) GetTemplate(ctx context.Context, tenantID uuid.UUID, templateID string) (*models.NotificationTemplate, error) {
	cacheKey := fmt.Sprintf("notification_template:%s:%s", tenantID.String(), templateID)
	data, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to get cached template: %v", err)
	}

	var tmpl models.NotificationTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %v", err)
	}
	return &tmpl, nil
}

func (s *notificationService) ListTemplates(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.NotificationTemplate, error) {
	// In production, this would query the database with proper indexing
	// For now, return empty slice as placeholder
	return []*models.NotificationTemplate{}, nil
}

// Configuration management methods
func (s *notificationService) UpdateNotificationConfig(ctx context.Context, tenantID uuid.UUID, config *models.NotificationConfig) error {
	config.UpdatedAt = time.Now()
	cacheKey := fmt.Sprintf("notification_config:%s:%s", tenantID.String(), config.Type)
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal notification config: %v", err)
	}

	return s.redisClient.Set(ctx, cacheKey, data, time.Hour).Err()
}

func (s *notificationService) GetNotificationConfig(ctx context.Context, tenantID uuid.UUID, notificationType models.NotificationType) (*models.NotificationConfig, error) {
	cacheKey := fmt.Sprintf("notification_config:%s:%s", tenantID.String(), notificationType)
	data, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("notification config not found")
		}
		return nil, fmt.Errorf("failed to get cached notification config: %v", err)
	}

	var config models.NotificationConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notification config: %v", err)
	}
	return &config, nil
}

// Alert management methods
func (s *notificationService) CreateAlert(ctx context.Context, tenantID uuid.UUID, alert *models.Alert) error {
	alert.ID = uuid.NewString()
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()

	// Process the alert by triggering notifications
	return s.ProcessAlert(ctx, tenantID, alert.ID)
}

func (s *notificationService) UpdateAlertStatus(ctx context.Context, tenantID uuid.UUID, alertID string, status string) error {
	cacheKey := fmt.Sprintf("alert:%s:%s", tenantID.String(), alertID)
	data, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("alert not found")
		}
		return fmt.Errorf("failed to get alert: %v", err)
	}

	var alert models.Alert
	if err := json.Unmarshal(data, &alert); err != nil {
		return fmt.Errorf("failed to unmarshal alert: %v", err)
	}

	alert.Status = status
	alert.UpdatedAt = time.Now()

	if status == "acknowledged" {
		now := time.Now()
		alert.AcknowledgeAt = &now
	}

	alertData, err := json.Marshal(&alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %v", err)
	}

	return s.redisClient.Set(ctx, cacheKey, alertData, time.Hour*24).Err()
}

func (s *notificationService) ProcessAlert(ctx context.Context, tenantID uuid.UUID, alertID string) error {
	// In production, get alert from database/cache and process notifications
	log.Printf("Processing alert %s for tenant %s", alertID, tenantID.String())
	return nil
}

func (s *notificationService) CheckAndTriggerAlerts(ctx context.Context, tenantID uuid.UUID) error {
	log.Printf("Checking and triggering alerts for tenant %s", tenantID.String())

	// Example alert triggers - in production these would be configurable
	alerts := []models.Alert{
		{
			AlertType: models.AlertTypeLowStock,
			Message:   "Low stock alert triggered",
			Data:      map[string]interface{}{"threshold": 10},
		},
	}

	for _, alert := range alerts {
		if err := s.CreateAlert(ctx, tenantID, &alert); err != nil {
			log.Printf("Failed to create alert: %v", err)
		}
	}

	return nil
}

// Webhook subscription management methods
func (s *notificationService) CreateWebhookSubscription(ctx context.Context, tenantID uuid.UUID, subscription *models.WebhookSubscription) error {
	subscription.ID = uuid.NewString()
	subscription.TenantID = tenantID.String()
	subscription.CreatedAt = time.Now()
	subscription.UpdatedAt = time.Now()

	cacheKey := fmt.Sprintf("webhook_subscription:%s:%s", tenantID.String(), subscription.ID)
	data, err := json.Marshal(subscription)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook subscription: %v", err)
	}

	return s.redisClient.Set(ctx, cacheKey, data, time.Hour).Err()
}

func (s *notificationService) UpdateWebhookSubscription(ctx context.Context, tenantID uuid.UUID, subscription *models.WebhookSubscription) error {
	subscription.UpdatedAt = time.Now()

	cacheKey := fmt.Sprintf("webhook_subscription:%s:%s", tenantID.String(), subscription.ID)
	data, err := json.Marshal(subscription)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook subscription: %v", err)
	}

	return s.redisClient.Set(ctx, cacheKey, data, time.Hour).Err()
}

func (s *notificationService) DeleteWebhookSubscription(ctx context.Context, tenantID uuid.UUID, subscriptionID string) error {
	cacheKey := fmt.Sprintf("webhook_subscription:%s:%s", tenantID.String(), subscriptionID)
	return s.redisClient.Del(ctx, cacheKey).Err()
}

func (s *notificationService) ListWebhookSubscriptions(ctx context.Context, tenantID uuid.UUID) ([]*models.WebhookSubscription, error) {
	// In production, this would query the database
	// For now, return empty slice as placeholder
	return []*models.WebhookSubscription{}, nil
}

// Alert configuration methods
func (s *notificationService) UpdateAlertConfig(ctx context.Context, tenantID uuid.UUID, config *models.AlertConfig) error {
	config.UpdatedAt = time.Now()

	cacheKey := fmt.Sprintf("alert_config:%s:%s", tenantID.String(), config.AlertType)
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal alert config: %v", err)
	}

	return s.redisClient.Set(ctx, cacheKey, data, time.Hour).Err()
}

func (s *notificationService) GetAlertConfig(ctx context.Context, tenantID uuid.UUID, alertType models.AlertType) (*models.AlertConfig, error) {
	cacheKey := fmt.Sprintf("alert_config:%s:%s", tenantID.String(), alertType)
	data, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("alert config not found")
		}
		return nil, fmt.Errorf("failed to get cached alert config: %v", err)
	}

	var config models.AlertConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal alert config: %v", err)
	}
	return &config, nil
}

// Utility methods
func (s *notificationService) RenderTemplate(tmplParam *models.NotificationTemplate, data map[string]interface{}) (string, error) {
	templateCacheKey := fmt.Sprintf("%s:%s", tmplParam.TenantID, tmplParam.ID)

	if tmpl, exists := s.templates[templateCacheKey]; exists {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("failed to execute template: %v", err)
		}
		return buf.String(), nil
	}

	tmpl, err := template.New(templateCacheKey).Parse(tmplParam.BodyTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %v", err)
	}

	s.templates[templateCacheKey] = tmpl

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}

func (s *notificationService) RetryFailedNotifications(ctx context.Context) error {
	// Placeholder for retrying failed notifications
	log.Println("Retrying failed notifications")
	return nil
}

// Cache template for faster rendering
func (s *notificationService) cacheTemplate(tmplParam *models.NotificationTemplate) {
	templateCacheKey := fmt.Sprintf("%s:%s", tmplParam.TenantID, tmplParam.ID)
	tmpl, err := template.New(templateCacheKey).Parse(tmplParam.BodyTemplate)
	if err != nil {
		log.Printf("Failed to cache template %s: %v", templateCacheKey, err)
		return
	}
	s.templates[templateCacheKey] = tmpl
}

// Helper methods
func (s *notificationService) getWebhookSubscription(ctx context.Context, tenantID uuid.UUID, subscriptionID string) (*models.WebhookSubscription, error) {
	cacheKey := fmt.Sprintf("webhook_subscription:%s:%s", tenantID.String(), subscriptionID)
	data, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("webhook subscription not found")
		}
		return nil, fmt.Errorf("failed to get webhook subscription: %v", err)
	}

	var subscription models.WebhookSubscription
	if err := json.Unmarshal(data, &subscription); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook subscription: %v", err)
	}
	return &subscription, nil
}

func (s *notificationService) updateWebhookSubscription(ctx context.Context, tenantID uuid.UUID, subscription *models.WebhookSubscription) error {
	cacheKey := fmt.Sprintf("webhook_subscription:%s:%s", tenantID.String(), subscription.ID)
	data, err := json.Marshal(subscription)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook subscription: %v", err)
	}

	return s.redisClient.Set(ctx, cacheKey, data, time.Hour).Err()
}