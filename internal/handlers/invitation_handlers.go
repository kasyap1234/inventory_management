package handlers

import (
	"net/http"

	"agromart2/internal/common"
	"agromart2/internal/services"
	"agromart2/internal/validation"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type InvitationHandlers struct {
	invitationService services.InvitationService
}

func NewInvitationHandlers(invitationService services.InvitationService) *InvitationHandlers {
	return &InvitationHandlers{invitationService: invitationService}
}

type CreateInviteRequest struct {
	Email       string   `json:"email" validate:"required,email"`
	RoleID      string   `json:"role_id" validate:"required,uuid4"`
	Permissions []string `json:"permissions"`
}

type AcceptInviteRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Password  string `json:"password" validate:"required,min=8"`
}

func (h *InvitationHandlers) CreateInvitation(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant context missing")
	}

	var req CreateInviteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	roleID, _ := uuid.Parse(req.RoleID)

	serviceReq := &services.CreateInvitationRequest{
		TenantID:    tenantID,
		Email:       req.Email,
		RoleID:      roleID,
		Permissions: req.Permissions,
	}

	userID, ok := common.GetUserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User context missing")
	}

	invitation, err := h.invitationService.CreateInvitation(ctx, serviceReq, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, invitation)
}

func (h *InvitationHandlers) GetInvitation(c echo.Context) error {
	ctx := c.Request().Context()
	token := c.Param("token")

	invitation, err := h.invitationService.GetInvitationByToken(ctx, token)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, invitation)
}

func (h *InvitationHandlers) AcceptInvitation(c echo.Context) error {
	ctx := c.Request().Context()
	token := c.Param("token")

	var req AcceptInviteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	validation.SanitizeStruct(&req)

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Validate password strength
	if err := common.ValidatePassword(req.Password); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	serviceReq := &services.AcceptInvitationRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  req.Password,
	}

	user, err := h.invitationService.AcceptInvitation(ctx, token, serviceReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Invitation accepted successfully",
		"user":    user,
	})
}

func (h *InvitationHandlers) ListInvitations(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant context missing")
	}

	// Pagination would go here
	invitations, err := h.invitationService.ListInvitations(ctx, tenantID, 100, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, invitations)
}

func (h *InvitationHandlers) RevokeInvitation(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid invitation ID")
	}

	if err := h.invitationService.RevokeInvitation(ctx, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Invitation revoked"})
}
