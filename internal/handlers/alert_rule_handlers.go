package handlers

import (
	"net/http"
	"strings"

	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AlertRuleHandlers handles alert rule HTTP requests
type AlertRuleHandlers struct {
	alertRuleService services.AlertRuleService
	rbacMiddleware   *middleware.RBACMiddleware
}

// NewAlertRuleHandlers creates a new alert rule handlers instance
func NewAlertRuleHandlers(alertRuleService services.AlertRuleService, rbacMiddleware *middleware.RBACMiddleware) *AlertRuleHandlers {
	return &AlertRuleHandlers{
		alertRuleService: alertRuleService,
		rbacMiddleware:   rbacMiddleware,
	}
}

// CreateAlertRule creates a new alert rule
func (h *AlertRuleHandlers) CreateAlertRule(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		Name        string                 `json:"name" validate:"required"`
		Description *string                `json:"description"`
		EventType   string                 `json:"event_type" validate:"required"`
		Conditions  map[string]interface{} `json:"conditions" validate:"required"`
		Actions     []models.AlertAction   `json:"actions" validate:"required"`
		IsActive    bool                   `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	rule := &models.AlertRule{
		ID:         uuid.New(),
		Name:       strings.TrimSpace(req.Name),
		EventType:  strings.TrimSpace(req.EventType),
		Conditions: req.Conditions,
		Actions:    req.Actions,
		IsActive:   req.IsActive,
	}

	if err := h.alertRuleService.CreateAlertRule(ctx, tenantID, rule); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, rule)
}

// UpdateAlertRule updates an existing alert rule
func (h *AlertRuleHandlers) UpdateAlertRule(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	ruleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid alert rule ID")
	}

	existing, err := h.alertRuleService.GetAlertRule(ctx, tenantID, ruleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Alert rule not found")
	}

	var req struct {
		Name        *string                `json:"name"`
		Description *string                `json:"description"`
		EventType   *string                `json:"event_type"`
		Conditions  map[string]interface{} `json:"conditions"`
		Actions     []models.AlertAction   `json:"actions"`
		IsActive    *bool                  `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if req.Name != nil {
		existing.Name = strings.TrimSpace(*req.Name)
	}
	if req.EventType != nil {
		existing.EventType = strings.TrimSpace(*req.EventType)
	}
	if req.Conditions != nil {
		existing.Conditions = req.Conditions
	}
	if req.Actions != nil {
		existing.Actions = req.Actions
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := h.alertRuleService.UpdateAlertRule(ctx, tenantID, existing); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, existing)
}

// GetAlertRule gets a specific alert rule by ID
func (h *AlertRuleHandlers) GetAlertRule(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	ruleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid alert rule ID")
	}

	rule, err := h.alertRuleService.GetAlertRule(ctx, tenantID, ruleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Alert rule not found")
	}

	return c.JSON(http.StatusOK, rule)
}

// ListAlertRules lists alert rules for a tenant
func (h *AlertRuleHandlers) ListAlertRules(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	eventType := c.QueryParam("event_type")

	rules, err := h.alertRuleService.ListAlertRules(ctx, tenantID, eventType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"alert_rules": rules,
		"count":       len(rules),
	})
}

// DeleteAlertRule deletes an alert rule
func (h *AlertRuleHandlers) DeleteAlertRule(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	ruleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid alert rule ID")
	}

	if err := h.alertRuleService.DeleteAlertRule(ctx, tenantID, ruleID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Alert rule deleted successfully",
	})
}

// TestAlertRule tests an alert rule with provided data
func (h *AlertRuleHandlers) TestAlertRule(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	ruleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid alert rule ID")
	}

	var req struct {
		TestData map[string]interface{} `json:"test_data" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	result, err := h.alertRuleService.TestAlertRule(ctx, tenantID, ruleID, req.TestData)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// EvaluateAlertRules manually triggers alert rule evaluation
func (h *AlertRuleHandlers) EvaluateAlertRules(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		EventType string                 `json:"event_type" validate:"required"`
		Data      map[string]interface{} `json:"data" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := h.alertRuleService.EvaluateAlertRules(ctx, tenantID, req.EventType, req.Data); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Alert rules evaluated successfully",
	})
}
