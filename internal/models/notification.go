package models

import (
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
	NotificationTypeWebhook NotificationType = "webhook"
)

// AlertType represents different types of alerts
type AlertType string

const (
	AlertTypeLowStock      AlertType = "low_stock"
	AlertTypeOrderIssue    AlertType = "order_issue"
	AlertTypeJobFailure    AlertType = "job_failure"
	AlertTypeInvoiceOverdue AlertType = "invoice_overdue"
)

// NotificationTemplate represents configurable notification templates
type NotificationTemplate struct {
	ID          string    `json:"id" db:"id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Type        string    `json:"type" db:"type"` // email, sms, webhook
	EventType   string    `json:"event_type" db:"event_type"`
	Subject     *string   `json:"subject" db:"subject"`
	BodyTemplate string    `json:"body_template" db:"body_template"`
	Variables   JSONB     `json:"variables" db:"variables"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// NotificationConfig represents tenant-specific notification configuration
type NotificationConfig struct {
	ID             string                 `json:"id" db:"id"`
	TenantID       string                 `json:"tenant_id" db:"tenant_id"`
	Type           NotificationType       `json:"type" db:"type"`
	Configuration  JSONB                  `json:"configuration" db:"configuration"`
	IsActive       bool                   `json:"is_active" db:"is_active"`
	WebhookTimeout *int                   `json:"webhook_timeout" db:"webhook_timeout"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// AlertConfig represents alert triggers configuration
type AlertConfig struct {
	ID               string     `json:"id" db:"id"`
	TenantID         string     `json:"tenant_id" db:"tenant_id"`
	AlertType        AlertType  `json:"alert_type" db:"alert_type"`
	Config           JSONB      `json:"config" db:"config"` // type-specific config
	Enabled          bool       `json:"enabled" db:"enabled"`
	NotificationChannels []NotificationType `json:"notification_channels" db:"-"` // from database array
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// Notification represents a sent notification
type Notification struct {
	ID                string          `json:"id" db:"id"`
	TenantID          string          `json:"tenant_id" db:"tenant_id"`
	Type              NotificationType `json:"type" db:"type"`
	EventType         string          `json:"event_type" db:"event_type"`
	EventID           string          `json:"event_id" db:"event_id"`
	Recipient         string          `json:"recipient" db:"recipient"`
	Subject           *string         `json:"subject" db:"subject"`
	Body              string          `json:"body" db:"body"`
	Status            string          `json:"status" db:"status"`
	Error             *string         `json:"error" db:"error"`
	ResponseData      JSONB           `json:"response_data" db:"response_data"`
	SentAt            *time.Time      `json:"sent_at" db:"sent_at"`
	RetryCount        int             `json:"retry_count" db:"retry_count"`
	MaxRetries        int             `json:"max_retries" db:"max_retries"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
}

// WebhookSubscription represents external webhook subscriptions
type WebhookSubscription struct {
	ID          string     `json:"id" db:"id"`
	TenantID    string     `json:"tenant_id" db:"tenant_id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description" db:"description"`
	URL         string     `json:"url" db:"url"`
	Secret      string     `json:"secret" db:"secret"`
	Events      []string   `json:"events" db:"-"` // from database array
	IsActive    bool       `json:"is_active" db:"is_active"`
	LastUsedAt  *time.Time `json:"last_used_at" db:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Alert represents an alert instance
type Alert struct {
	ID               string                 `json:"id" db:"id"`
	TenantID         string                 `json:"tenant_id" db:"tenant_id"`
	AlertType        AlertType              `json:"alert_type" db:"alert_type"`
	EventID          string                 `json:"event_id" db:"event_id"`
	Message          string                 `json:"message" db:"message"`
	Data             JSONB                  `json:"data" db:"data"`
	Severity         string                 `json:"severity" db:"severity"`
	Status           string                 `json:"status" db:"status"` // pending, sent, acknowledged
	NotificationIDs  []string               `json:"notification_ids" db:"-"` // from database array
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
	AcknowledgeAt    *time.Time             `json:"acknowledget_at" db:"acknowledged_at"`
}

// JSONB represents PostgreSQL JSONB type
type JSONB map[string]interface{}

// LowStockAlertData represents data for low stock alert
type LowStockAlertData struct {
	ProductID      string  `json:"product_id"`
	ProductName    string  `json:"product_name"`
	WarehouseID    string  `json:"warehouse_id"`
	WarehouseName  string  `json:"warehouse_name"`
	CurrentStock   int     `json:"current_stock"`
	MinimumStock   int     `json:"minimum_stock"`
	Threshold      int     `json:"threshold"`
}

// OrderIssueAlertData represents data for order issue alert
type OrderIssueAlertData struct {
	OrderID         string  `json:"order_id"`
	OrderType       string  `json:"order_type"`
	Status          string  `json:"status"`
	DaysOverdue     *int    `json:"days_overdue"`
	IssueType       string  `json:"issue_type"`
}

// JobFailureAlertData represents data for job failure alert
type JobFailureAlertData struct {
	JobName         string  `json:"job_name"`
	JobID           string  `json:"job_id"`
	ErrorMessage    string  `json:"error_message"`
	RetryCount      int     `json:"retry_count"`
}

// InvoiceOverdueAlertData represents data for invoice overdue alert
type InvoiceOverdueAlertData struct {
	InvoiceID       string  `json:"invoice_id"`
	OrderID         string  `json:"order_id"`
	DueDate         string  `json:"due_date"`
	DaysOverdue     int     `json:"days_overdue"`
	Amount          float64 `json:"amount"`
}