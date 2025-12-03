package handlers

import (
	"net/http"

	"agromart2/internal/common"
	"agromart2/internal/services"

	"github.com/labstack/echo/v4"
)

type SuperAdminHandlers struct {
	invitationService services.InvitationService
}

func NewSuperAdminHandlers(invitationService services.InvitationService) *SuperAdminHandlers {
	return &SuperAdminHandlers{invitationService: invitationService}
}

type InviteTenantAdminRequest struct {
	Email      string `json:"email" validate:"required,email"`
	TenantName string `json:"tenant_name" validate:"required"`
}

func (h *SuperAdminHandlers) InviteTenantAdmin(c echo.Context) error {
	ctx := c.Request().Context()

	// Ensure user is super admin (this should be handled by middleware, but good to double check or rely on middleware)
	// We assume the route is protected by "system.admin" permission.

	userID, ok := common.GetUserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User context missing")
	}

	var req InviteTenantAdminRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	invitation, err := h.invitationService.InviteTenantAdmin(ctx, req.Email, req.TenantName, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, invitation)
}
