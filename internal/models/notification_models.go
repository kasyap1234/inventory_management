package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification channel
type NotificationType string

const (
	NotificationTypeEmail   NotificationType = "email"
	NotificationTypeSMS     NotificationType = "sms"
	NotificationTypeWebhook NotificationType = "webhook"
	NotificationTypePush    NotificationType = "push"
)

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	TenantID     uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name         string                 `json:"name" db:"name"`
	Type         NotificationType       `json:"type" db:"type"`
	EventType    string                 `json:"event_type" db:"event_type"`
	Subject      *string                `json:"subject" db:"subject"`
	BodyTemplate string                 `json:"body_template" db:"body_template"`
	Variables    map[string]interface{} `json:"variables" db:"variables"`
	IsActive     bool                   `json:"is_active" db:"is_active"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// WebhookSubscription represents a webhook subscription
type WebhookSubscription struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	TenantID       uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name           string                 `json:"name" db:"name"`
	URL            string                 `json:"url" db:"url"`
	Secret         *string                `json:"secret,omitempty" db:"secret"`
	EventTypes     []string               `json:"event_types" db:"event_types"`
	Headers        map[string]interface{} `json:"headers" db:"headers"`
	TimeoutSeconds int                    `json:"timeout_seconds" db:"timeout_seconds"`
	RetryCount     int                    `json:"retry_count" db:"retry_count"`
	IsActive       bool                   `json:"is_active" db:"is_active"`
	LastSuccessAt  *time.Time             `json:"last_success_at" db:"last_success_at"`
	LastFailureAt  *time.Time             `json:"last_failure_at" db:"last_failure_at"`
	FailureCount   int                    `json:"failure_count" db:"failure_count"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// AlertRule represents an alert rule configuration
type AlertRule struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	TenantID        uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name            string                 `json:"name" db:"name"`
	EventType       string                 `json:"event_type" db:"event_type"`
	Conditions      map[string]interface{} `json:"conditions" db:"conditions"`
	Actions         []AlertAction          `json:"actions" db:"actions"`
	IsActive        bool                   `json:"is_active" db:"is_active"`
	LastTriggeredAt *time.Time             `json:"last_triggered_at" db:"last_triggered_at"`
	TriggerCount    int                    `json:"trigger_count" db:"trigger_count"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// AlertAction represents an action to take when an alert is triggered
type AlertAction struct {
	Type       string                 `json:"type"`   // email, sms, webhook
	Target     string                 `json:"target"` // recipient or webhook ID
	TemplateID *string                `json:"template_id"`
	CustomData map[string]interface{} `json:"custom_data"`
}

// NotificationConfig represents user notification preferences
type NotificationConfig struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	EventType string    `json:"event_type" db:"event_type"`
	Channels  []string  `json:"channels" db:"channels"`
	IsEnabled bool      `json:"is_enabled" db:"is_enabled"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NotificationDelivery represents a notification delivery attempt
type NotificationDelivery struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	TenantID       uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	NotificationID *uuid.UUID             `json:"notification_id" db:"notification_id"`
	TemplateID     *uuid.UUID             `json:"template_id" db:"template_id"`
	WebhookID      *uuid.UUID             `json:"webhook_id" db:"webhook_id"`
	Channel        NotificationType       `json:"channel" db:"channel"`
	Recipient      string                 `json:"recipient" db:"recipient"`
	Status         string                 `json:"status" db:"status"`
	AttemptCount   int                    `json:"attempt_count" db:"attempt_count"`
	LastAttemptAt  *time.Time             `json:"last_attempt_at" db:"last_attempt_at"`
	DeliveredAt    *time.Time             `json:"delivered_at" db:"delivered_at"`
	ErrorMessage   *string                `json:"error_message" db:"error_message"`
	ResponseData   map[string]interface{} `json:"response_data" db:"response_data"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// Enhanced Notification model extending the basic notification
type EnhancedNotification struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	TenantID         uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	UserID           uuid.UUID              `json:"user_id" db:"user_id"`
	Title            string                 `json:"title" db:"title"`
	Message          string                 `json:"message" db:"message"`
	NotificationType string                 `json:"notification_type" db:"notification_type"`
	Priority         string                 `json:"priority" db:"priority"`
	Status           string                 `json:"status" db:"status"`
	EventType        *string                `json:"event_type" db:"event_type"`
	EventData        map[string]interface{} `json:"event_data" db:"event_data"`
	TemplateID       *uuid.UUID             `json:"template_id" db:"template_id"`
	ReadAt           *time.Time             `json:"read_at" db:"read_at"`
	ExpiresAt        *time.Time             `json:"expires_at" db:"expires_at"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// NotificationEvent represents an event that can trigger notifications
type NotificationEvent struct {
	Type      string                 `json:"type"`
	TenantID  uuid.UUID              `json:"tenant_id"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// Template variable validation
type TemplateVariable struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // string, number, boolean, date
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
}

// Validation methods

// ValidateTemplate validates a notification template
func (nt *NotificationTemplate) ValidateTemplate() error {
	if nt.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if nt.EventType == "" {
		return fmt.Errorf("event type is required")
	}
	if nt.BodyTemplate == "" {
		return fmt.Errorf("body template is required")
	}
	if nt.Type == NotificationTypeEmail && nt.Subject == nil {
		return fmt.Errorf("subject is required for email templates")
	}
	return nil
}

// ValidateWebhook validates a webhook subscription
func (ws *WebhookSubscription) ValidateWebhook() error {
	if ws.Name == "" {
		return fmt.Errorf("webhook name is required")
	}
	if ws.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}
	if len(ws.EventTypes) == 0 {
		return fmt.Errorf("at least one event type is required")
	}
	if ws.TimeoutSeconds <= 0 || ws.TimeoutSeconds > 300 {
		return fmt.Errorf("timeout must be between 1 and 300 seconds")
	}
	if ws.RetryCount < 0 || ws.RetryCount > 10 {
		return fmt.Errorf("retry count must be between 0 and 10")
	}
	return nil
}

// ValidateAlertRule validates an alert rule
func (ar *AlertRule) ValidateAlertRule() error {
	if ar.Name == "" {
		return fmt.Errorf("alert rule name is required")
	}
	if ar.EventType == "" {
		return fmt.Errorf("event type is required")
	}
	if len(ar.Conditions) == 0 {
		return fmt.Errorf("at least one condition is required")
	}
	if len(ar.Actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}
	return nil
}

// JSON marshaling helpers for database storage

// MarshalVariables marshals variables to JSON for database storage
func (nt *NotificationTemplate) MarshalVariables() ([]byte, error) {
	if nt.Variables == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(nt.Variables)
}

// UnmarshalVariables unmarshals variables from JSON
func (nt *NotificationTemplate) UnmarshalVariables(data []byte) error {
	if len(data) == 0 {
		nt.Variables = make(map[string]interface{})
		return nil
	}
	return json.Unmarshal(data, &nt.Variables)
}

// MarshalEventTypes marshals event types to JSON for database storage
func (ws *WebhookSubscription) MarshalEventTypes() ([]byte, error) {
	return json.Marshal(ws.EventTypes)
}

// UnmarshalEventTypes unmarshals event types from JSON
func (ws *WebhookSubscription) UnmarshalEventTypes(data []byte) error {
	return json.Unmarshal(data, &ws.EventTypes)
}

// MarshalHeaders marshals headers to JSON for database storage
func (ws *WebhookSubscription) MarshalHeaders() ([]byte, error) {
	if ws.Headers == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(ws.Headers)
}

// UnmarshalHeaders unmarshals headers from JSON
func (ws *WebhookSubscription) UnmarshalHeaders(data []byte) error {
	if len(data) == 0 {
		ws.Headers = make(map[string]interface{})
		return nil
	}
	return json.Unmarshal(data, &ws.Headers)
}
