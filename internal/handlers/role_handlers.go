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

// RoleHandlers handles role-related HTTP requests
type RoleHandlers struct {
	rbacService  services.RBACService
	rbacMiddleware *middleware.RBACMiddleware
}

// NewRoleHandlers creates a new role handlers instance
func NewRoleHandlers(rbacService services.RBACService, rbacMiddleware *middleware.RBACMiddleware) *RoleHandlers {
	return &RoleHandlers{
		rbacService:  rbacService,
		rbacMiddleware: rbacMiddleware,
	}
}

// CreateRoleRequest represents request body for creating a role
type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// UpdateRoleRequest represents request body for updating a role
type UpdateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// RoleResponse represents response body for a role
type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

// ListRolesRequest represents query parameters for listing roles
type ListRolesRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

// toRoleResponse converts a model.Role to a RoleResponse
func toRoleResponse(role *models.Role) RoleResponse {
	return RoleResponse{
		ID:          role.ID,
		TenantID:    role.TenantID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// CreateRole handles creating a new role
func (h *RoleHandlers) CreateRole(c echo.Context) error {
	err := h.rbacMiddleware.RequirePermission("roles:create")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Get tenant from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	role := &models.Role{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.rbacService.CreateRole(ctx, tenantID, role); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create role")
	}

	return c.JSON(http.StatusCreated, toRoleResponse(role))
}

// ListRoles handles getting a list of roles
func (h *RoleHandlers) ListRoles(c echo.Context) error {
	err := h.rbacMiddleware.RequirePermission("roles:read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Get tenant from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req ListRolesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	roles, err := h.rbacService.ListRoles(ctx, tenantID, req.Limit, req.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list roles")
	}

	response := make([]RoleResponse, len(roles))
	for i, role := range roles {
		response[i] = toRoleResponse(role)
	}

	return c.JSON(http.StatusOK, response)
}

// GetRole handles getting a single role by ID
func (h *RoleHandlers) GetRole(c echo.Context) error {
	err := h.rbacMiddleware.RequirePermission("roles:read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Get tenant from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	role, err := h.rbacService.GetRoleByID(ctx, tenantID, roleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get role")
	}

	if role == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Role not found")
	}

	return c.JSON(http.StatusOK, toRoleResponse(role))
}

// UpdateRole handles updating an existing role
func (h *RoleHandlers) UpdateRole(c echo.Context) error {
	err := h.rbacMiddleware.RequirePermission("roles:update")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Get tenant from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	var req UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get existing role
	role, err := h.rbacService.GetRoleByID(ctx, tenantID, roleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get role")
	}

	if role == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Role not found")
	}

	// Update role fields
	role.Name = req.Name
	role.Description = req.Description

	if err := h.rbacService.UpdateRole(ctx, tenantID, role); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update role")
	}

	return c.JSON(http.StatusOK, toRoleResponse(role))
}

// DeleteRole handles deleting a role
func (h *RoleHandlers) DeleteRole(c echo.Context) error {
	err := h.rbacMiddleware.RequirePermission("roles:delete")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Get tenant from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	// Check if role exists
	role, err := h.rbacService.GetRoleByID(ctx, tenantID, roleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get role")
	}

	if role == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Role not found")
	}

	if err := h.rbacService.DeleteRole(ctx, tenantID, roleID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete role")
	}

	return c.NoContent(http.StatusNoContent)
}
