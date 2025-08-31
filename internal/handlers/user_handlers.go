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

// UserHandlers handles user-related HTTP requests
type UserHandlers struct {
	userRepo       repositories.UserRepository
	tenantRepo     repositories.TenantRepository
	rbacMiddleware *middleware.RBACMiddleware
}

// NewUserHandlers creates a new user handlers instance
func NewUserHandlers(userRepo repositories.UserRepository, tenantRepo repositories.TenantRepository, rbacMiddleware *middleware.RBACMiddleware) *UserHandlers {
	return &UserHandlers{
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
		rbacMiddleware: rbacMiddleware,
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
	err := h.rbacMiddleware.RequirePermission("users:list")(func(c echo.Context) error {
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
	Email     string  `json:"email" validate:"required,email"`
	FirstName string  `json:"first_name" validate:"required"`
	LastName  string  `json:"last_name" validate:"required"`
	TenantID  string  `json:"tenant_id" validate:"required"`
	Status    *string `json:"status"`
}

// CreateUser handles creating a new user
func (h *UserHandlers) CreateUser(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("users:create")(func(c echo.Context) error {
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

	c.Logger().Infof("DEBUG: Requested tenant_id from body: %s", req.TenantID)
	c.Logger().Infof("DEBUG: Parsed tenant_id UUID: %s", tenantID.String())

	// Check if this is a cross-tenant operation
	currentUserTenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found for current user")
	}

	isCrossTenantOperation := tenantID != currentUserTenantID
	if isCrossTenantOperation {
		c.Logger().Infof("DEBUG: Cross-tenant operation detected - requested: %s vs user: %s",
			tenantID.String(), currentUserTenantID.String())

		// Check for admin permissions
		err := h.rbacMiddleware.RequirePermission("users:create_any_tenant")(func(c echo.Context) error { return nil })(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions for cross-tenant user creation")
		}

		// Set explicit tenant override for JWT middleware
		c.Set("explicit_tenant_id", tenantID)
		c.Logger().Infof("DEBUG: Cross-tenant operation granted - explicit_tenant_id set for middleware override")
	} else {
		c.Logger().Infof("DEBUG: Same-tenant operation - no override needed")
	}

	// Validate that tenant exists
	tenant, err := h.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant does not exist")
	}
	if tenant == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant does not exist")
	}

	c.Logger().Infof("DEBUG: Tenant validation successful: %s (ID: %s)", tenant.Name, tenant.ID.String())

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

	c.Logger().Infof("DEBUG: User object before repository.Create:")
	c.Logger().Infof("DEBUG:   User ID: %s", user.ID.String())
	c.Logger().Infof("DEBUG:   User Tenant ID: %s (expected: %s)", user.TenantID.String(), tenantID.String())
	c.Logger().Infof("DEBUG:   User Email: %s", user.Email)

	if err := h.userRepo.Create(ctx, user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	c.Logger().Infof("DEBUG: User creation successful, retrieving created user to verify tenant association...")
	createdUser, err := h.userRepo.GetByID(ctx, tenantID, userID)
	if err == nil && createdUser != nil {
		c.Logger().Infof("DEBUG: Created user retrieved - Actual tenant_id: %s", createdUser.TenantID.String())
		if createdUser.TenantID != tenantID {
			c.Logger().Errorf("CRITICAL BUG: Created user has wrong tenant_id! Expected: %s, Actual: %s", tenantID.String(), createdUser.TenantID.String())
		}
	} else {
		c.Logger().Errorf("DEBUG: Could not retrieve created user for verification")
	}

	return c.JSON(http.StatusCreated, user)
}

// GetUser handles getting user details by ID
func (h *UserHandlers) GetUser(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("users:read")(func(c echo.Context) error {
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
	err := h.rbacMiddleware.RequirePermission("users:update")(func(c echo.Context) error {
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
	err := h.rbacMiddleware.RequirePermission("users:delete")(func(c echo.Context) error {
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