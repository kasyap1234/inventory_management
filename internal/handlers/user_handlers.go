package handlers

import (
	"fmt"
	"net/http"

	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// UserHandlers handles user-related HTTP requests
type UserHandlers struct {
	userRepo       repositories.UserRepository
	tenantRepo     repositories.TenantRepository
	rbacMiddleware *middleware.RBACMiddleware
	userService    services.UserService
	roleService    services.RoleManagementService
}

// NewUserHandlers creates a new user handlers instance
func NewUserHandlers(userRepo repositories.UserRepository, tenantRepo repositories.TenantRepository, rbacMiddleware *middleware.RBACMiddleware, userService services.UserService, roleService services.RoleManagementService) *UserHandlers {
	return &UserHandlers{
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
		rbacMiddleware: rbacMiddleware,
		userService:    userService,
		roleService:    roleService,
	}
}

// ListUsersRequest represents query parameters for listing users
type ListUsersRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

// ListUsers handles getting a list of users with tenant filtering
func (h *UserHandlers) ListUsers(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("user.list")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err // RBAC middleware will return appropriate error
	}

	ctx := c.Request().Context()

	var req ListUsersRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100 // Maximum limit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get users from the tenant
	users, err := h.userRepo.List(ctx, tenantID, req.Limit, req.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list users")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"users":  users,
		"limit":  req.Limit,
		"offset": req.Offset,
	})
}

// CreateUserRequest represents the user creation request payload
type CreateUserRequest struct {
	Email       string   `json:"email" validate:"required,email"`
	FirstName   string   `json:"first_name" validate:"required"`
	LastName    string   `json:"last_name" validate:"required"`
	TenantID    string   `json:"tenant_id" validate:"required"`
	Status      *string  `json:"status"`
	RoleID      *string  `json:"role_id"`
	Permissions []string `json:"permissions"`
}

// CreateUser handles creating a new user
func (h *UserHandlers) CreateUser(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("user.create")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if req.Email == "" || req.FirstName == "" || req.LastName == "" || req.TenantID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email, first name, last name, and tenant ID are required")
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid tenant ID format")
	}
	if tenantID == uuid.Nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid tenant ID")
	}

	// Check if this is a cross-tenant operation
	currentUserTenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found for current user")
	}

	isCrossTenantOperation := tenantID != currentUserTenantID
	if isCrossTenantOperation {

		// Check for admin permissions
		err := h.rbacMiddleware.RequirePermission("user.create_any_tenant")(func(c echo.Context) error { return nil })(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions for cross-tenant user creation")
		}

		// Set explicit tenant override for JWT middleware
		c.Set("explicit_tenant_id", tenantID)
	}

	// Validate that tenant exists
	tenant, err := h.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant does not exist")
	}
	if tenant == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant does not exist")
	}

	// Check if user already exists in this tenant
	existingUser, err := h.userRepo.GetByEmail(ctx, tenantID, req.Email)
	if err == nil && existingUser != nil {
		return echo.NewHTTPError(http.StatusConflict, "User already exists")
	}

	// Set default status if not provided
	status := "active"
	if req.Status != nil {
		status = *req.Status
	}

	// Generate user ID
	userID := uuid.New()

	// Create new user
	user := &models.User{
		ID:        userID,
		TenantID:  tenantID,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Status:    status,
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	// Handle Role/Permission Assignment
	if len(req.Permissions) > 0 {
		// 1. Create a custom role for this user
		roleName := fmt.Sprintf("custom_role_%s_%d", req.Email, 12345) // Simple unique name
		// Better unique name using timestamp would require time import, keeping it simple for now or adding time import if needed.
		// Actually, let's just use a UUID suffix or similar if possible, or just timestamp.
		// I'll assume time is imported or I'll add it.
		// Wait, I can't easily add imports with multi_replace if they are far away.
		// I'll use a hardcoded suffix for now or rely on something else.
		// Actually, I'll just use "custom_" + uuid.New().String()
		roleName = fmt.Sprintf("custom_%s", uuid.New().String())

		roleDesc := fmt.Sprintf("Custom role for %s", req.Email)
		role := &models.Role{
			Name:        roleName,
			Description: &roleDesc,
			IsActive:    true,
		}
		if err := h.roleService.CreateRole(ctx, tenantID, role); err != nil {
			// Log error but don't fail user creation? Or fail?
			// Better to fail or warn.
			return echo.NewHTTPError(http.StatusInternalServerError, "User created but failed to create custom role: "+err.Error())
		}

		// 2. Resolve Permission Names to IDs
		allPerms, err := h.roleService.ListAvailablePermissions(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list permissions")
		}

		var permIDs []uuid.UUID
		permMap := make(map[string]uuid.UUID)
		for _, p := range allPerms {
			permMap[p.Name] = p.ID
		}

		for _, pName := range req.Permissions {
			if id, ok := permMap[pName]; ok {
				permIDs = append(permIDs, id)
			}
		}

		// 3. Assign permissions to the new role
		if len(permIDs) > 0 {
			if err := h.roleService.AssignPermissionsToRole(ctx, tenantID, role.ID, permIDs); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to assign permissions to custom role")
			}
		}

		// 4. Assign the new role to the user
		if err := h.roleService.AssignUserToRole(ctx, tenantID, userID, role.ID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to assign custom role to user")
		}

	} else if req.RoleID != nil {
		// Assign specific role
		roleID, err := uuid.Parse(*req.RoleID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid Role ID")
		}
		if err := h.roleService.AssignUserToRole(ctx, tenantID, userID, roleID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to assign role to user")
		}
	} else {
		// Assign default 'user' role
		// We need to find the role named 'user' (or 'viewer'/'operator' based on seeds?)
		// The seeds created 'admin', 'manager', 'operator', 'viewer'.
		// 'user' role was mentioned in auth_service.go but seeds use specific names.
		// Let's try to find 'operator' as a safe default if 'user' doesn't exist.
		// Or better, let's look for 'user' first as auth_service creates it.
		roles, err := h.roleService.ListRoles(ctx, tenantID)
		if err == nil {
			var defaultRoleID uuid.UUID
			for _, r := range roles {
				if r.Name == "user" {
					defaultRoleID = r.ID
					break
				}
			}
			// If 'user' not found, try 'viewer'
			if defaultRoleID == uuid.Nil {
				for _, r := range roles {
					if r.Name == "viewer" {
						defaultRoleID = r.ID
						break
					}
				}
			}

			if defaultRoleID != uuid.Nil {
				h.roleService.AssignUserToRole(ctx, tenantID, userID, defaultRoleID)
			}
		}
	}

	return c.JSON(http.StatusCreated, user)
}

// GetUser handles getting user details by ID
func (h *UserHandlers) GetUser(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("user.read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	userIDStr := c.Param("id")
	if userIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User ID is required")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID format")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get user details
	user, err := h.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	return c.JSON(http.StatusOK, user)
}

// UpdateUserRequest represents the user update request payload
type UpdateUserRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Status    *string `json:"status"`
}

// UpdateUser handles updating user details
func (h *UserHandlers) UpdateUser(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("user.update")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	userIDStr := c.Param("id")
	if userIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User ID is required")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID format")
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get existing user
	user, err := h.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	// Update fields if provided
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := h.userRepo.Update(ctx, user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user")
	}

	return c.JSON(http.StatusOK, user)
}

// DeleteUserRequest represents the user deletion request payload (may include confirmation)
type DeleteUserRequest struct {
	Force *bool `json:"force"` // Force delete even if user has dependencies
}

// DeleteUser handles deleting a user
func (h *UserHandlers) DeleteUser(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("user.delete")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	userIDStr := c.Param("id")
	if userIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User ID is required")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID format")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req DeleteUserRequest
	if err := c.Bind(&req); err != nil {
		// Bind is optional for delete, but we'll proceed
		req.Force = nil
	}

	// Optional: Check if user exists before deleting
	_, err = h.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	if err := h.userRepo.Delete(ctx, tenantID, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete user")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

// UpdateUserProfileRequest represents the user profile update request payload
type UpdateUserProfileRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

// UpdateUserProfile handles updating the current user's profile
func (h *UserHandlers) UpdateUserProfile(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT context
	userID, ok := common.GetUserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	var req UpdateUserProfileRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "First name and last name are required")
	}

	// Call the service to update the profile
	if err := h.userService.UpdateUserProfile(ctx, userID, req.FirstName, req.LastName); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user profile")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Profile updated successfully",
	})
}

// GetPendingUsers handles listing users pending approval
func (h *UserHandlers) GetPendingUsers(c echo.Context) error {
	// Use RBAC middleware directly - require admin access
	err := h.rbacMiddleware.RequirePermission("user.approve")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	var req ListUsersRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get pending users
	users, err := h.userRepo.ListByStatus(ctx, tenantID, "pending_approval", req.Limit, req.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list pending users")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"users":  users,
		"limit":  req.Limit,
		"offset": req.Offset,
	})
}

// ApproveUser handles approving a pending user
func (h *UserHandlers) ApproveUser(c echo.Context) error {
	// Use RBAC middleware directly - require admin access
	err := h.rbacMiddleware.RequirePermission("user.approve")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	userIDStr := c.Param("id")
	if userIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User ID is required")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID format")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get existing user
	user, err := h.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	if user.Status != "pending_approval" {
		return echo.NewHTTPError(http.StatusBadRequest, "User is not pending approval")
	}

	// Update status to active
	if err := h.userRepo.UpdateStatus(ctx, tenantID, userID, "active"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to approve user")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "User approved successfully",
	})
}

// GetDirectory returns a directory of users for the tenant (simplified view)
func (h *UserHandlers) GetDirectory(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Pagination
	limit := 100
	offset := 0

	users, err := h.userRepo.List(ctx, tenantID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch directory")
	}

	// Filter sensitive info
	var directory []map[string]interface{}
	for _, u := range users {
		directory = append(directory, map[string]interface{}{
			"id":         u.ID,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"email":      u.Email,
			// Add role info if needed
		})
	}

	return c.JSON(http.StatusOK, directory)
}
