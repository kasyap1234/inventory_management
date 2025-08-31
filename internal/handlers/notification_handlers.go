package handlers

import (
	"fmt"
	"net/http"

	"agromart2/internal/middleware"
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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

	subscription := &models.WebhookSubscription{
		Name:        req.Name,
		Description: req.Description,
		URL:         req.URL,
		Secret:      req.Secret,
		Events:      req.Events,
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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

// DeleteWebhookSubscription deletes a webhook subscription
func (h *NotificationHandlers) DeleteWebhookSubscription(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		Type        string                 `json:"type" validate:"required"`
		EventType   string                 `json:"event_type" validate:"required"`
		Subject     *string                `json:"subject"`
		BodyTemplate string                `json:"body_template" validate:"required"`
		Variables   map[string]interface{} `json:"variables"`
		IsActive    bool                   `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	template := &models.NotificationTemplate{
		Type:         req.Type,
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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

	config := &models.NotificationConfig{
		Type:           req.Type,
		Configuration:  req.Configuration,
		IsActive:       req.IsActive,
		WebhookTimeout: req.WebhookTimeout,
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		AlertType        models.AlertType        `json:"alert_type" validate:"required"`
		Config           map[string]interface{}  `json:"config" validate:"required"`
		Enabled          bool                   `json:"enabled"`
		NotificationChannels []models.NotificationType `json:"notification_channels"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	config := &models.AlertConfig{
		AlertType:         req.AlertType,
		Config:           req.Config,
		Enabled:          req.Enabled,
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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
		"template_id":    req.TemplateID,
		"rendered_body": rendered,
		"subject":       template.Subject,
	})
}