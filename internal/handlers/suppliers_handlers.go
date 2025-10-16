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

// SupplierHandlers handles supplier-related HTTP requests
type SupplierHandlers struct {
	supplierService services.SupplierService
	rbacMiddleware  *middleware.RBACMiddleware
}

// NewSupplierHandlers creates a new supplier handlers instance
func NewSupplierHandlers(supplierService services.SupplierService, rbacMiddleware *middleware.RBACMiddleware) *SupplierHandlers {
	return &SupplierHandlers{
		supplierService: supplierService,
		rbacMiddleware:  rbacMiddleware,
	}
}

// ListSuppliersRequest represents query parameters for listing suppliers
type ListSuppliersRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

// ListSuppliers handles getting a list of suppliers with tenant filtering
func (h *SupplierHandlers) ListSuppliers(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("suppliers:list")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err // RBAC middleware will return appropriate error
	}

	ctx := c.Request().Context()

	var req ListSuppliersRequest
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

	// Get suppliers from the tenant
	suppliers, err := h.supplierService.List(ctx, tenantID, req.Limit, req.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list suppliers")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"suppliers": suppliers,
		"limit":     req.Limit,
		"offset":    req.Offset,
	})
}

// CreateSupplierRequest represents the supplier creation request payload
type CreateSupplierRequest struct {
	Name           string  `json:"name" validate:"required"`
	ContactEmail   *string `json:"contact_email"`
	ContactPhone   *string `json:"contact_phone"`
	Address        *string `json:"address"`
	LicenseNumber  *string `json:"license_number"`
}

// CreateSupplier handles creating a new supplier
func (h *SupplierHandlers) CreateSupplier(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("suppliers:create")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	var req CreateSupplierRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Name is required")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Create new supplier
	supplier := &models.Supplier{
		Name:          req.Name,
		ContactEmail:  req.ContactEmail,
		ContactPhone:  req.ContactPhone,
		Address:       req.Address,
		LicenseNumber: req.LicenseNumber,
	}

	if err := h.supplierService.Create(ctx, tenantID, supplier); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, supplier)
}

// GetSupplier handles getting supplier details by ID
func (h *SupplierHandlers) GetSupplier(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("suppliers:read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	supplierIDStr := c.Param("id")
	if supplierIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Supplier ID is required")
	}

	supplierID, err := uuid.Parse(supplierIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid supplier ID format")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get supplier details
	supplier, err := h.supplierService.GetByID(ctx, tenantID, supplierID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Supplier not found")
	}

	return c.JSON(http.StatusOK, supplier)
}

// UpdateSupplierRequest represents the supplier update request payload
type UpdateSupplierRequest struct {
	Name          *string `json:"name"`
	ContactEmail  *string `json:"contact_email"`
	ContactPhone  *string `json:"contact_phone"`
	Address       *string `json:"address"`
	LicenseNumber *string `json:"license_number"`
}

// UpdateSupplier handles updating supplier details
func (h *SupplierHandlers) UpdateSupplier(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("suppliers:update")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	supplierIDStr := c.Param("id")
	if supplierIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Supplier ID is required")
	}

	supplierID, err := uuid.Parse(supplierIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid supplier ID format")
	}

	var req UpdateSupplierRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get existing supplier
	supplier, err := h.supplierService.GetByID(ctx, tenantID, supplierID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Supplier not found")
	}

	// Update fields if provided
	if req.Name != nil {
		supplier.Name = *req.Name
	}
	if req.ContactEmail != nil {
		supplier.ContactEmail = req.ContactEmail
	}
	if req.ContactPhone != nil {
		supplier.ContactPhone = req.ContactPhone
	}
	if req.Address != nil {
		supplier.Address = req.Address
	}
	if req.LicenseNumber != nil {
		supplier.LicenseNumber = req.LicenseNumber
	}

	if err := h.supplierService.Update(ctx, tenantID, supplier); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, supplier)
}

// DeleteSupplier handles deleting a supplier
func (h *SupplierHandlers) DeleteSupplier(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("suppliers:delete")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	supplierIDStr := c.Param("id")
	if supplierIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Supplier ID is required")
	}

	supplierID, err := uuid.Parse(supplierIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid supplier ID format")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Check if supplier exists before deleting
	_, err = h.supplierService.GetByID(ctx, tenantID, supplierID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Supplier not found")
	}

	if err := h.supplierService.Delete(ctx, tenantID, supplierID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete supplier")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Supplier deleted successfully",
	})
}