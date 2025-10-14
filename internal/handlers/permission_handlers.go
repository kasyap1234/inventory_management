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

// PermissionHandlers handles permission-related HTTP requests
type PermissionHandlers struct {
	rbacService  services.RBACService
	rbacMiddleware *middleware.RBACMiddleware
}

// NewPermissionHandlers creates a new permission handlers instance
func NewPermissionHandlers(rbacService services.RBACService, rbacMiddleware *middleware.RBACMiddleware) *PermissionHandlers {
	return &PermissionHandlers{
		rbacService:  rbacService,
		rbacMiddleware: rbacMiddleware,
	}
}

// CreatePermissionRequest represents request body for creating a permission
type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// UpdatePermissionRequest represents request body for updating a permission
type UpdatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// PermissionResponse represents response body for a permission
type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   string    `json:"created_at"`
}

// ListPermissionsRequest represents query parameters for listing permissions
type ListPermissionsRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

// AssignPermissionRequest represents request body for assigning permissions to a role
type AssignPermissionRequest struct {
	PermissionIDs []uuid.UUID `json:"permission_ids" validate:"required,min=1"`
}

// toPermissionResponse converts a model.Permission to a PermissionResponse
func toPermissionResponse(permission *models.Permission) PermissionResponse {
	return PermissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ListPermissions handles getting a list of permissions
func (h *PermissionHandlers) ListPermissions(c echo.Context) error {
	err := h.rbacMiddleware.RequirePermission("permissions:read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	var req ListPermissionsRequest
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

	permissions, err := h.rbacService.ListPermissions(ctx, req.Limit, req.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list permissions")
	}

	response := make([]PermissionResponse, len(permissions))
	for i, permission := range permissions {
		response[i] = toPermissionResponse(permission)
	}

	return c.JSON(http.StatusOK, response)
}

// AssignPermissionsToRole handles assigning permissions to a role
func (h *PermissionHandlers) AssignPermissionsToRole(c echo.Context) error {
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

	var req AssignPermissionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Verify role exists and belongs to tenant
	role, err := h.rbacService.GetRoleByID(ctx, tenantID, roleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify role")
	}

	if role == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Role not found")
	}

	// Assign each permission to the role
	for _, permissionID := range req.PermissionIDs {
		if err := h.rbacService.AssignPermissionToRole(ctx, tenantID, roleID, permissionID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to assign permission")
		}
	}

	return c.NoContent(http.StatusNoContent)
}

// RevokePermissionFromRole handles revoking a permission from a role
func (h *PermissionHandlers) RevokePermissionFromRole(c echo.Context) error {
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

	permissionID, err := uuid.Parse(c.Param("permissionId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid permission ID")
	}

	// Verify role exists and belongs to tenant
	role, err := h.rbacService.GetRoleByID(ctx, tenantID, roleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify role")
	}

	if role == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Role not found")
	}

	if err := h.rbacService.RevokePermissionFromRole(ctx, tenantID, roleID, permissionID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to revoke permission")
	}

	return c.NoContent(http.StatusNoContent)
}

// GetRolePermissions handles getting permissions assigned to a role
func (h *PermissionHandlers) GetRolePermissions(c echo.Context) error {
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

	// Verify role exists and belongs to tenant
	role, err := h.rbacService.GetRoleByID(ctx, tenantID, roleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify role")
	}

	if role == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Role not found")
	}

	permissions, err := h.rbacService.GetRolePermissions(ctx, tenantID, roleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get role permissions")
	}

	response := make([]PermissionResponse, len(permissions))
	for i, permission := range permissions {
		response[i] = toPermissionResponse(permission)
	}

	return c.JSON(http.StatusOK, response)
}
