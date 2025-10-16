package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/services"
	"github.com/labstack/echo/v4"
)

// NotificationHandlers handles notification-related HTTP requests
type NotificationHandlers struct {
	notificationSvc services.NotificationService
}

// NewNotificationHandlers creates a new notification handlers instance
func NewNotificationHandlers(notificationSvc services.NotificationService) *NotificationHandlers {
	return &NotificationHandlers{
		notificationSvc: notificationSvc,
	}
}

// SendNotification sends a notification
func (h *NotificationHandlers) SendNotification(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		Type      models.NotificationType `json:"type" validate:"required"`
		EventType string                  `json:"event_type" validate:"required"`
		EventID   string                  `json:"event_id"`
		Recipient string                  `json:"recipient" validate:"required"`
		Subject   *string                 `json:"subject"`
		Body      string                  `json:"body" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	notification := &models.Notification{
		Type:       req.Type,
		EventType:  req.EventType,
		EventID:    req.EventID,
		Recipient:  req.Recipient,
		Subject:    req.Subject,
		Body:       req.Body,
		SentAt:     nil,
		RetryCount: 0,
		MaxRetries: 3,
	}

	if err := h.notificationSvc.SendNotification(ctx, tenantID, notification); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Notification sent successfully",
	})
}

// CreateWebhookSubscription creates a webhook subscription
func (h *NotificationHandlers) CreateWebhookSubscription(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

    var req struct {
        Name        string   `json:"name" validate:"required"`
        Description *string  `json:"description"`
        URL         string   `json:"url" validate:"required,url"`
        Secret      string   `json:"secret" validate:"required"`
        Events      []string `json:"events" validate:"required"`
        IsActive    bool     `json:"is_active"`
    }

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	secret := strings.TrimSpace(req.Secret)
    subscription := &models.WebhookSubscription{
		Name:        strings.TrimSpace(req.Name),
		URL:         strings.TrimSpace(req.URL),
		Secret:      &secret,
		EventTypes:  req.Events,
		IsActive:    req.IsActive,
	}

	if err := h.notificationSvc.CreateWebhookSubscription(ctx, tenantID, subscription); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, subscription)
}

// ListWebhookSubscriptions lists webhook subscriptions for a tenant
func (h *NotificationHandlers) ListWebhookSubscriptions(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	subscriptions, err := h.notificationSvc.ListWebhookSubscriptions(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"webhook_subscriptions": subscriptions,
	})
}

// UpdateWebhookSubscription updates an existing webhook subscription
func (h *NotificationHandlers) UpdateWebhookSubscription(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	subscriptionID := c.Param("id")
	if strings.TrimSpace(subscriptionID) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Subscription ID is required")
	}

    existing, err := h.notificationSvc.GetWebhookSubscription(ctx, tenantID, subscriptionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

    var req struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		URL         *string  `json:"url"`
		Secret      *string  `json:"secret"`
		Events      []string `json:"events"`
		IsActive    *bool    `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if req.Name != nil {
		existing.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		// Note: WebhookSubscription does not have a Description field
		// If you need to add one, update the WebhookSubscription model
	}
	if req.URL != nil {
		trimmed := strings.TrimSpace(*req.URL)
		if trimmed != "" {
			existing.URL = trimmed
		}
	}
	if req.Secret != nil {
		trimmed := strings.TrimSpace(*req.Secret)
		if trimmed != "" {
			existing.Secret = &trimmed  // Secret is a *string field
		}
	}
	if len(req.Events) > 0 {
		existing.EventTypes = req.Events  // Field is EventTypes, not Events
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := h.notificationSvc.UpdateWebhookSubscription(ctx, tenantID, existing); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, existing)
}

// DeleteWebhookSubscription deletes a webhook subscription
func (h *NotificationHandlers) DeleteWebhookSubscription(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Subscription ID is required")
	}

	if err := h.notificationSvc.DeleteWebhookSubscription(ctx, tenantID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Webhook subscription deleted successfully",
	})
}

// CreateTemplate creates a notification template
func (h *NotificationHandlers) CreateTemplate(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

    var req struct {
		Type         string                 `json:"type" validate:"required"`
		EventType    string                 `json:"event_type" validate:"required"`
		Subject      *string                `json:"subject"`
		BodyTemplate string                 `json:"body_template" validate:"required"`
		Variables    map[string]interface{} `json:"variables"`
		IsActive     bool                   `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

    template := &models.NotificationTemplate{
		Name:         req.EventType, // Use event type as name for now
		Type:         models.NotificationType(req.Type),
		EventType:    req.EventType,
		Subject:      req.Subject,
		BodyTemplate: req.BodyTemplate,
		Variables:    req.Variables,
		IsActive:     req.IsActive,
	}

	if err := h.notificationSvc.CreateTemplate(ctx, tenantID, template); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, template)
}

// ListTemplates lists notification templates for a tenant
func (h *NotificationHandlers) ListTemplates(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	eventType := c.QueryParam("event_type")

    templates, err := h.notificationSvc.ListTemplates(ctx, tenantID, eventType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"templates": templates,
	})
}

// DeleteTemplate deletes a notification template
func (h *NotificationHandlers) DeleteTemplate(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Template ID is required")
	}

    if err := h.notificationSvc.DeleteTemplate(ctx, tenantID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Template deleted successfully",
	})
}

// UpdateNotificationConfig updates notification configuration
func (h *NotificationHandlers) UpdateNotificationConfig(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

    var req struct {
        Type           models.NotificationType `json:"type" validate:"required"`
        Configuration  map[string]interface{}  `json:"configuration" validate:"required"`
        IsActive       bool                    `json:"is_active"`
        WebhookTimeout *int                    `json:"webhook_timeout"`
    }

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

    // NotificationConfig structure per models/notification_models.go:
    // EventType string, Channels []string, IsEnabled bool
    config := &models.NotificationConfig{
		EventType:  string(req.Type),  // Convert NotificationType to string
		IsEnabled:  true,  // Default to enabled
		Channels:   []string{"email"},  // Default channels
	}

	if err := h.notificationSvc.UpdateNotificationConfig(ctx, tenantID, config); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Notification configuration updated successfully",
	})
}

// UpdateAlertConfig updates alert configuration
func (h *NotificationHandlers) UpdateAlertConfig(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

    var req struct {
		AlertType            models.AlertType          `json:"alert_type" validate:"required"`
		Config               map[string]interface{}    `json:"config" validate:"required"`
		Enabled              bool                      `json:"enabled"`
		NotificationChannels []models.NotificationType `json:"notification_channels"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	config := &models.AlertConfig{
		AlertType: req.AlertType,
		Config:    req.Config,
		Enabled:   req.Enabled,
	}

	if err := h.notificationSvc.UpdateAlertConfig(ctx, tenantID, config); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Alert configuration updated successfully",
	})
}

// TriggerAlerts manually triggers alert checks
func (h *NotificationHandlers) TriggerAlerts(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	if err := h.notificationSvc.CheckAndTriggerAlerts(ctx, tenantID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Alert check and trigger completed",
	})
}

// RenderTemplate tests template rendering with provided data
func (h *NotificationHandlers) RenderTemplate(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

    var req struct {
		TemplateID string                 `json:"template_id" validate:"required"`
		Data       map[string]interface{} `json:"data" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	template, err := h.notificationSvc.GetTemplate(ctx, tenantID, req.TemplateID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Template not found")
	}

	if !template.IsActive {
		return echo.NewHTTPError(http.StatusBadRequest, "Template is not active")
	}

    rendered, err := h.notificationSvc.RenderTemplate(template, req.Data)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to render template: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"template_id":   req.TemplateID,
		"rendered_body": rendered,
		"subject":       template.Subject,
	})
}

// ListNotifications lists notifications for a tenant
func (h *NotificationHandlers) ListNotifications(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Parse query parameters for filtering
	notificationType := c.QueryParam("type")
	eventType := c.QueryParam("event_type")
	status := c.QueryParam("status") // e.g., "pending", "sent", "failed"

	notifications, err := h.notificationSvc.ListNotifications(ctx, tenantID, notificationType, eventType, status)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// GetNotification gets a specific notification by ID
func (h *NotificationHandlers) GetNotification(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Notification ID is required")
	}

	notification, err := h.notificationSvc.GetNotification(ctx, tenantID, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Notification not found")
	}

	return c.JSON(http.StatusOK, notification)
}

// MarkNotificationAsRead marks a notification as read
func (h *NotificationHandlers) MarkNotificationAsRead(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Notification ID is required")
	}

	if err := h.notificationSvc.MarkAsRead(ctx, tenantID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Notification marked as read",
	})
}

// MarkAllNotificationsAsRead marks all notifications as read for the tenant
func (h *NotificationHandlers) MarkAllNotificationsAsRead(c echo.Context) error {
    ctx := c.Request().Context()

    tenantID, ok := common.GetTenantIDFromContext(ctx)
    if !ok {
        return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
    }

    if err := h.notificationSvc.MarkAllAsRead(ctx, tenantID); err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }

    return c.JSON(http.StatusOK, map[string]string{
        "message": "All notifications marked as read",
    })
}

// ArchiveNotification archives a single notification
func (h *NotificationHandlers) ArchiveNotification(c echo.Context) error {
    ctx := c.Request().Context()

    tenantID, ok := common.GetTenantIDFromContext(ctx)
    if !ok {
        return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
    }

    id := c.Param("id")
    if id == "" {
        return echo.NewHTTPError(http.StatusBadRequest, "Notification ID is required")
    }

    if err := h.notificationSvc.ArchiveNotification(ctx, tenantID, id); err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }

    return c.JSON(http.StatusOK, map[string]string{
        "message": "Notification archived successfully",
    })
}

// DeleteNotification deletes a notification
func (h *NotificationHandlers) DeleteNotification(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Notification ID is required")
	}

	if err := h.notificationSvc.DeleteNotification(ctx, tenantID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Notification deleted successfully",
	})
}
