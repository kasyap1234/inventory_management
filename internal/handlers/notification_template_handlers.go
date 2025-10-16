package handlers

import (
	"net/http"

	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type NotificationTemplateHandlers struct {
	templateService services.NotificationTemplateService
	rbacMiddleware  *middleware.RBACMiddleware
}

func NewNotificationTemplateHandlers(
	templateService services.NotificationTemplateService,
	rbacMiddleware *middleware.RBACMiddleware,
) *NotificationTemplateHandlers {
	return &NotificationTemplateHandlers{
		templateService: templateService,
		rbacMiddleware:  rbacMiddleware,
	}
}

// ListTemplates returns all notification templates for a tenant
func (h *NotificationTemplateHandlers) ListTemplates(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	eventType := c.QueryParam("event_type")

	templates, err := h.templateService.ListTemplates(ctx, tenantID, eventType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch templates")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":   "Templates retrieved successfully",
		"templates": templates,
	})
}

// GetTemplate returns a specific notification template
func (h *NotificationTemplateHandlers) GetTemplate(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid template ID")
	}

	template, err := h.templateService.GetTemplate(ctx, tenantID, templateID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Template not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":  "Template retrieved successfully",
		"template": template,
	})
}

// CreateTemplate creates a new notification template
func (h *NotificationTemplateHandlers) CreateTemplate(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		Name         string                 `json:"name" validate:"required"`
		Type         models.NotificationType `json:"type" validate:"required"`
		EventType    string                 `json:"event_type" validate:"required"`
		Subject      *string                `json:"subject"`
		BodyTemplate string                 `json:"body_template" validate:"required"`
		Variables    map[string]interface{} `json:"variables"`
		IsActive     bool                   `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	template := &models.NotificationTemplate{
		TenantID:     tenantID,
		Name:         req.Name,
		Type:         req.Type,
		EventType:    req.EventType,
		Subject:      req.Subject,
		BodyTemplate: req.BodyTemplate,
		Variables:    req.Variables,
		IsActive:     req.IsActive,
	}

	if err := h.templateService.CreateTemplate(ctx, template); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create template")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message":  "Template created successfully",
		"template": template,
	})
}

// UpdateTemplate updates an existing notification template
func (h *NotificationTemplateHandlers) UpdateTemplate(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid template ID")
	}

	var req struct {
		Name         string                 `json:"name" validate:"required"`
		Type         models.NotificationType `json:"type" validate:"required"`
		EventType    string                 `json:"event_type" validate:"required"`
		Subject      *string                `json:"subject"`
		BodyTemplate string                 `json:"body_template" validate:"required"`
		Variables    map[string]interface{} `json:"variables"`
		IsActive     bool                   `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	template, err := h.templateService.GetTemplate(ctx, tenantID, templateID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Template not found")
	}

	template.Name = req.Name
	template.Type = req.Type
	template.EventType = req.EventType
	template.Subject = req.Subject
	template.BodyTemplate = req.BodyTemplate
	template.Variables = req.Variables
	template.IsActive = req.IsActive

	if err := h.templateService.UpdateTemplate(ctx, template); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update template")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":  "Template updated successfully",
		"template": template,
	})
}

// DeleteTemplate deletes a notification template
func (h *NotificationTemplateHandlers) DeleteTemplate(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid template ID")
	}

	if err := h.templateService.DeleteTemplate(ctx, tenantID, templateID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete template")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Template deleted successfully",
	})
}

// TestTemplate tests a notification template with sample data
func (h *NotificationTemplateHandlers) TestTemplate(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid template ID")
	}

	var req struct {
		TestData map[string]interface{} `json:"test_data" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result, err := h.templateService.TestTemplate(ctx, tenantID, templateID, req.TestData)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to test template: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Template tested successfully",
		"result":  result,
	})
}
