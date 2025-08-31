package handlers

import (
	"net/http"
	"strconv"

	"agromart2/internal/middleware"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// SubscriptionHandlers handles HTTP requests for subscriptions
type SubscriptionHandlers struct {
	subscriptionService services.SubscriptionService
	rbacMiddleware      *middleware.RBACMiddleware
}

// NewSubscriptionHandlers creates a new subscription handlers instance
func NewSubscriptionHandlers(subscriptionService services.SubscriptionService, rbacMiddleware *middleware.RBACMiddleware) *SubscriptionHandlers {
	return &SubscriptionHandlers{
		subscriptionService: subscriptionService,
		rbacMiddleware:      rbacMiddleware,
	}
}

// validateUUID validates UUID string
func (h *SubscriptionHandlers) validateUUID(idStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}
	return id, nil
}

// CreateSubscription handles POST /subscriptions
func (h *SubscriptionHandlers) CreateSubscription(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		PlanID        string `json:"plan_id" validate:"required"`
		CustomerEmail string `json:"customer_email" validate:"required,email"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if req.PlanID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Plan ID is required")
	}

	if req.CustomerEmail == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Customer email is required")
	}

	subscription, err := h.subscriptionService.Create(ctx, tenantID, req.PlanID, req.CustomerEmail)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message":      "Subscription created successfully",
		"subscription": subscription,
	})
}

// ListSubscriptions handles GET /subscriptions
func (h *SubscriptionHandlers) ListSubscriptions(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	limit := 10  // default
	offset := 0  // default

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	subscriptions, err := h.subscriptionService.List(ctx, tenantID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"subscriptions": subscriptions,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetSubscriptionByID handles GET /subscriptions/:id
func (h *SubscriptionHandlers) GetSubscriptionByID(c echo.Context) error {
	ctx := c.Request().Context()

	subscriptionID, err := h.validateUUID(c.Param("id"))
	if err != nil {
		return err
	}

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	subscription, err := h.subscriptionService.GetByID(ctx, tenantID, subscriptionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, subscription)
}

// CancelSubscription handles DELETE /subscriptions/:id/cancel
func (h *SubscriptionHandlers) CancelSubscription(c echo.Context) error {
	ctx := c.Request().Context()

	subscriptionID, err := h.validateUUID(c.Param("id"))
	if err != nil {
		return err
	}

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	err = h.subscriptionService.Cancel(ctx, tenantID, subscriptionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Subscription cancelled successfully",
	})
}

// PauseSubscription handles PUT /subscriptions/:id/pause
func (h *SubscriptionHandlers) PauseSubscription(c echo.Context) error {
	ctx := c.Request().Context()

	subscriptionID, err := h.validateUUID(c.Param("id"))
	if err != nil {
		return err
	}

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	err = h.subscriptionService.Pause(ctx, tenantID, subscriptionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Subscription paused successfully",
	})
}

// ResumeSubscription handles PUT /subscriptions/:id/resume
func (h *SubscriptionHandlers) ResumeSubscription(c echo.Context) error {
	ctx := c.Request().Context()

	subscriptionID, err := h.validateUUID(c.Param("id"))
	if err != nil {
		return err
	}

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	err = h.subscriptionService.Resume(ctx, tenantID, subscriptionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Subscription resumed successfully",
	})
}

// UpdateSubscriptionPlan handles PUT /subscriptions/:id/plan
func (h *SubscriptionHandlers) UpdateSubscriptionPlan(c echo.Context) error {
	ctx := c.Request().Context()

	subscriptionID, err := h.validateUUID(c.Param("id"))
	if err != nil {
		return err
	}

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		PlanID string `json:"plan_id" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if req.PlanID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Plan ID is required")
	}

	err = h.subscriptionService.UpdatePlan(ctx, tenantID, subscriptionID, req.PlanID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Subscription plan updated successfully",
	})
}

// ValidateBilling handles GET /subscriptions/:id/validate-billing
func (h *SubscriptionHandlers) ValidateBilling(c echo.Context) error {
	ctx := c.Request().Context()

	subscriptionID, err := h.validateUUID(c.Param("id"))
	if err != nil {
		return err
	}

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	err = h.subscriptionService.ValidateBilling(ctx, tenantID, subscriptionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Billing validation successful",
	})
}

// GetAvailablePlans handles GET /subscriptions/plans
func (h *SubscriptionHandlers) GetAvailablePlans(c echo.Context) error {
	plans := h.subscriptionService.GetAvailablePlans()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"plans":               plans,
		"message":            "Available subscription plans retrieved successfully",
	})
}