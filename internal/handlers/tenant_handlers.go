package handlers

import (
	"net/http"

	"agromart2/internal/middleware"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// TenantHandlers handles tenant-related HTTP requests
type TenantHandlers struct {
	tenantService  services.TenantService
	rbacMiddleware *middleware.RBACMiddleware
}

// NewTenantHandlers creates a new tenant handlers instance
func NewTenantHandlers(tenantService services.TenantService, rbacMiddleware *middleware.RBACMiddleware) *TenantHandlers {
	return &TenantHandlers{
		tenantService:  tenantService,
		rbacMiddleware: rbacMiddleware,
	}
}

// ListTenantsRequest represents query parameters for listing tenants
type ListTenantsRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

// ListTenants handles getting a list of tenants (admin only)
func (h *TenantHandlers) ListTenants(c echo.Context) error {
	// Check admin permission - only admins can list all tenants
	err := h.rbacMiddleware.RequirePermission("tenant.list")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	var req ListTenantsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// List tenants
	tenants, err := h.tenantService.List(c.Request().Context(), req.Limit, req.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list tenants")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tenants": tenants,
		"limit":   req.Limit,
		"offset":  req.Offset,
	})
}

// CreateTenantRequest represents the tenant creation request payload
type CreateTenantRequest struct {
	Name       string `json:"name" validate:"required"`
	Subdomain  string `json:"subdomain" validate:"required"`
	License    string `json:"license" validate:"required"`
	AdminEmail string `json:"admin_email" validate:"required,email"`
}

// CreateTenant handles creating a new tenant (admin only)
func (h *TenantHandlers) CreateTenant(c echo.Context) error {
	var req CreateTenantRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if req.Name == "" || req.Subdomain == "" || req.License == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Name, subdomain, and license are required")
	}

	// Basic validation for subdomain format
	if len(req.Subdomain) < 3 {
		return echo.NewHTTPError(http.StatusBadRequest, "Subdomain must be at least 3 characters long")
	}

	// Create tenant request
	tenantReq := &services.CreateTenantRequest{
		Name:       req.Name,
		Subdomain:  req.Subdomain,
		License:    req.License,
		AdminEmail: req.AdminEmail,
	}

	// Create tenant
	tenant, err := h.tenantService.Create(c.Request().Context(), tenantReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create tenant")
	}

	return c.JSON(http.StatusCreated, tenant)
}

// GetTenant handles getting tenant details by ID
func (h *TenantHandlers) GetTenant(c echo.Context) error {
	// Check if user can read tenant details
	err := h.rbacMiddleware.RequirePermission("tenant.read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	tenantIDStr := c.Param("id")
	if tenantIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant ID is required")
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid tenant ID format")
	}

	// Get tenant details
	tenant, err := h.tenantService.GetByID(c.Request().Context(), tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Tenant not found")
	}

	return c.JSON(http.StatusOK, tenant)
}

// UpdateTenantRequest represents the tenant update request payload
type UpdateTenantRequest struct {
	Name      *string `json:"name"`
	Subdomain *string `json:"subdomain"`
	License   *string `json:"license"`
	Status    *string `json:"status"`
}

// UpdateTenant handles updating tenant details
func (h *TenantHandlers) UpdateTenant(c echo.Context) error {
	// Check admin permission
	err := h.rbacMiddleware.RequirePermission("tenant.update")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	tenantIDStr := c.Param("id")
	if tenantIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant ID is required")
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid tenant ID format")
	}

	var req UpdateTenantRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Get existing tenant first
	existing, err := h.tenantService.GetByID(c.Request().Context(), tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Tenant not found")
	}

	// Build update request with existing or new values
	updateReq := &services.UpdateTenantRequest{
		ID:        tenantID,
		Name:      existing.Name,      // Use existing value as default
		Subdomain: existing.Subdomain, // Use existing value as default
		License:   existing.License,   // Use existing value as default
		Status:    existing.Status,    // Use existing value as default
	}

	// Override with provided values if not nil
	if req.Name != nil {
		updateReq.Name = *req.Name
	}
	if req.Subdomain != nil {
		updateReq.Subdomain = *req.Subdomain
	}
	if req.License != nil {
		updateReq.License = *req.License
	}
	if req.Status != nil {
		updateReq.Status = *req.Status
	}

	// Update tenant
	if err := h.tenantService.Update(c.Request().Context(), updateReq); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update tenant")
	}

	// Return updated tenant
	updatedTenant, err := h.tenantService.GetByID(c.Request().Context(), tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Tenant updated but failed to retrieve")
	}

	return c.JSON(http.StatusOK, updatedTenant)
}

// DeleteTenantRequest represents the tenant deletion request payload
type DeleteTenantRequest struct {
	Force *bool `json:"force"` // Force delete even if tenant has dependencies
}

// DeleteTenant handles deleting a tenant (admin only)
func (h *TenantHandlers) DeleteTenant(c echo.Context) error {
	// Check admin permission
	err := h.rbacMiddleware.RequirePermission("tenant.delete")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	tenantIDStr := c.Param("id")
	if tenantIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant ID is required")
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid tenant ID format")
	}

	var req DeleteTenantRequest
	if err := c.Bind(&req); err != nil {
		// Bind is optional for delete, but we'll proceed
		req.Force = nil
	}

	// Check if tenant exists
	_, err = h.tenantService.GetByID(c.Request().Context(), tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Tenant not found")
	}

	// Delete tenant
	if err := h.tenantService.Delete(c.Request().Context(), tenantID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete tenant")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Tenant deleted successfully",
	})
}

// GetTenantBySubdomain handles getting tenant details by subdomain
func (h *TenantHandlers) GetTenantBySubdomain(c echo.Context) error {
	subdomain := c.Param("subdomain")
	if subdomain == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Subdomain is required")
	}

	// Get tenant by subdomain
	tenant, err := h.tenantService.GetBySubdomain(c.Request().Context(), subdomain)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Tenant not found")
	}

	return c.JSON(http.StatusOK, tenant)
}

type UpdateTenantSettingsRequest struct {
	Name      string `json:"name" validate:"required"`
	Subdomain string `json:"subdomain" validate:"required"`
	License   string `json:"license"`
}

func (h *TenantHandlers) GetTenantSettings(c echo.Context) error {
	tenantID, ok := c.Request().Context().Value("tenant_id").(uuid.UUID)
	if !ok || tenantID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant ID not found in context")
	}

	tenant, err := h.tenantService.GetTenantSettings(c.Request().Context(), tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Tenant settings not found")
	}

	return c.JSON(http.StatusOK, tenant)
}

func (h *TenantHandlers) UpdateTenantSettings(c echo.Context) error {
	tenantID, ok := c.Request().Context().Value("tenant_id").(uuid.UUID)
	if !ok || tenantID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant ID not found in context")
	}

	var req UpdateTenantSettingsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if req.Name == "" || req.Subdomain == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Name and subdomain are required")
	}

	updateReq := &services.UpdateTenantSettingsRequest{
		TenantID:  tenantID,
		Name:      req.Name,
		Subdomain: req.Subdomain,
		License:   req.License,
	}

	if err := h.tenantService.UpdateTenantSettings(c.Request().Context(), updateReq); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	updatedTenant, err := h.tenantService.GetTenantSettings(c.Request().Context(), tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Settings updated but failed to retrieve")
	}

	return c.JSON(http.StatusOK, updatedTenant)
}
