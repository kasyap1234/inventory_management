package handlers

import (
	"net/http"

	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type RoleHandlers struct {
	roleRepo           repositories.RoleRepository
	permissionRepo     repositories.PermissionRepository
	rolePermissionRepo repositories.RolePermissionRepository
	rbacMiddleware     *middleware.RBACMiddleware
}

func NewRoleHandlers(
	roleRepo repositories.RoleRepository,
	permissionRepo repositories.PermissionRepository,
	rolePermissionRepo repositories.RolePermissionRepository,
	rbacMiddleware *middleware.RBACMiddleware,
) *RoleHandlers {
	return &RoleHandlers{
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
		rolePermissionRepo: rolePermissionRepo,
		rbacMiddleware:     rbacMiddleware,
	}
}

// ListRoles returns all roles for a tenant
func (h *RoleHandlers) ListRoles(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	roles, err := h.roleRepo.List(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch roles")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Roles retrieved successfully",
		"roles":   roles,
	})
}

// GetRole returns a specific role
func (h *RoleHandlers) GetRole(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	role, err := h.roleRepo.GetByID(ctx, tenantID, roleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Role not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role retrieved successfully",
		"role":    role,
	})
}

// CreateRole creates a new role
func (h *RoleHandlers) CreateRole(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		Name        string  `json:"name" validate:"required"`
		Description *string `json:"description"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	role := &models.Role{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.roleRepo.Create(ctx, role); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create role")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Role created successfully",
		"role":    role,
	})
}

// UpdateRole updates an existing role
func (h *RoleHandlers) UpdateRole(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	var req struct {
		Name        string  `json:"name" validate:"required"`
		Description *string `json:"description"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	role, err := h.roleRepo.GetByID(ctx, tenantID, roleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Role not found")
	}

	role.Name = req.Name
	role.Description = req.Description

	if err := h.roleRepo.Update(ctx, role); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update role")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role updated successfully",
		"role":    role,
	})
}

// DeleteRole deletes a role
func (h *RoleHandlers) DeleteRole(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	if err := h.roleRepo.Delete(ctx, tenantID, roleID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete role")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role deleted successfully",
	})
}

// ListPermissions returns all available permissions
func (h *RoleHandlers) ListPermissions(c echo.Context) error {
	ctx := c.Request().Context()

	permissions, err := h.permissionRepo.List(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch permissions")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":     "Permissions retrieved successfully",
		"permissions": permissions,
	})
}

// GetRolePermissions returns permissions assigned to a role
func (h *RoleHandlers) GetRolePermissions(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	permissions, err := h.rolePermissionRepo.GetPermissionsByRole(ctx, tenantID, roleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch role permissions")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":     "Role permissions retrieved successfully",
		"permissions": permissions,
	})
}

// AssignPermissionsToRole assigns permissions to a role
func (h *RoleHandlers) AssignPermissionsToRole(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	var req struct {
		PermissionIDs []string `json:"permission_ids" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	permissionIDs := make([]uuid.UUID, 0, len(req.PermissionIDs))
	for _, idStr := range req.PermissionIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid permission ID: "+idStr)
		}
		permissionIDs = append(permissionIDs, id)
	}

	if err := h.rolePermissionRepo.RemoveAllPermissionsFromRole(ctx, tenantID, roleID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to clear existing permissions")
	}

	for _, permID := range permissionIDs {
		if err := h.rolePermissionRepo.AssignPermissionToRole(ctx, tenantID, roleID, permID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to assign permission")
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Permissions assigned successfully",
	})
}

// RemovePermissionFromRole removes a permission from a role
func (h *RoleHandlers) RemovePermissionFromRole(c echo.Context) error {
	ctx := c.Request().Context()

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

	if err := h.rolePermissionRepo.RemovePermissionFromRole(ctx, tenantID, roleID, permissionID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to remove permission")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Permission removed successfully",
	})
}

// GetUserRoles returns all roles assigned to a user
func (h *RoleHandlers) GetUserRoles(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	roles, err := h.roleRepo.GetUserRoles(ctx, tenantID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch user roles")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "User roles retrieved successfully",
		"roles":   roles,
	})
}

// AssignRolesToUser assigns roles to a user
func (h *RoleHandlers) AssignRolesToUser(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	var req struct {
		RoleIDs []string `json:"role_ids" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Remove all existing roles first
	existingRoles, err := h.roleRepo.GetUserRoles(ctx, tenantID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch existing roles")
	}

	for _, role := range existingRoles {
		if err := h.roleRepo.RemoveUserFromRole(ctx, tenantID, userID, role.ID); err != nil {
			// Continue even if removal fails
		}
	}

	// Assign new roles
	for _, roleIDStr := range req.RoleIDs {
		roleID, err := uuid.Parse(roleIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID: "+roleIDStr)
		}

		if err := h.roleRepo.AssignUserToRole(ctx, tenantID, userID, roleID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to assign role")
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Roles assigned successfully",
	})
}

// RemoveRoleFromUser removes a role from a user
func (h *RoleHandlers) RemoveRoleFromUser(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	if err := h.roleRepo.RemoveUserFromRole(ctx, tenantID, userID, roleID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to remove role from user")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role removed from user successfully",
	})
}

// GetRoleUsers returns all users assigned to a role
func (h *RoleHandlers) GetRoleUsers(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	users, err := h.roleRepo.GetRoleUsers(ctx, tenantID, roleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch role users")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role users retrieved successfully",
		"users":   users,
	})
}
