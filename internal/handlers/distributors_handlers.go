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

// DistributorHandlers handles distributor-related HTTP requests
type DistributorHandlers struct {
	distributorService services.DistributorService
	rbacMiddleware     *middleware.RBACMiddleware
}

// NewDistributorHandlers creates a new distributor handlers instance
func NewDistributorHandlers(distributorService services.DistributorService, rbacMiddleware *middleware.RBACMiddleware) *DistributorHandlers {
	return &DistributorHandlers{
		distributorService: distributorService,
		rbacMiddleware:     rbacMiddleware,
	}
}

// ListDistributorsRequest represents query parameters for listing distributors
type ListDistributorsRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

// ListDistributors handles getting a list of distributors with tenant filtering
func (h *DistributorHandlers) ListDistributors(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("distributors:list")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err // RBAC middleware will return appropriate error
	}

	ctx := c.Request().Context()

	var req ListDistributorsRequest
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

	// Get distributors from the tenant
	distributors, err := h.distributorService.List(ctx, tenantID, req.Limit, req.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list distributors")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"distributors": distributors,
		"limit":        req.Limit,
		"offset":       req.Offset,
	})
}

// CreateDistributorRequest represents the distributor creation request payload
type CreateDistributorRequest struct {
	Name           string  `json:"name" validate:"required"`
	ContactEmail   *string `json:"contact_email"`
	ContactPhone   *string `json:"contact_phone"`
	Address        *string `json:"address"`
	LicenseNumber  *string `json:"license_number"`
}

// CreateDistributor handles creating a new distributor
func (h *DistributorHandlers) CreateDistributor(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("distributors:create")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	var req CreateDistributorRequest
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

	// Create new distributor
	distributor := &models.Distributor{
		Name:          req.Name,
		ContactEmail:  req.ContactEmail,
		ContactPhone:  req.ContactPhone,
		Address:       req.Address,
		LicenseNumber: req.LicenseNumber,
	}

	if err := h.distributorService.Create(ctx, tenantID, distributor); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, distributor)
}

// GetDistributor handles getting distributor details by ID
func (h *DistributorHandlers) GetDistributor(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("distributors:read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	distributorIDStr := c.Param("id")
	if distributorIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Distributor ID is required")
	}

	distributorID, err := uuid.Parse(distributorIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid distributor ID format")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get distributor details
	distributor, err := h.distributorService.GetByID(ctx, tenantID, distributorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Distributor not found")
	}

	return c.JSON(http.StatusOK, distributor)
}

// UpdateDistributorRequest represents the distributor update request payload
type UpdateDistributorRequest struct {
	Name          *string `json:"name"`
	ContactEmail  *string `json:"contact_email"`
	ContactPhone  *string `json:"contact_phone"`
	Address       *string `json:"address"`
	LicenseNumber *string `json:"license_number"`
}

// UpdateDistributor handles updating distributor details
func (h *DistributorHandlers) UpdateDistributor(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("distributors:update")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	distributorIDStr := c.Param("id")
	if distributorIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Distributor ID is required")
	}

	distributorID, err := uuid.Parse(distributorIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid distributor ID format")
	}

	var req UpdateDistributorRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get existing distributor
	distributor, err := h.distributorService.GetByID(ctx, tenantID, distributorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Distributor not found")
	}

	// Update fields if provided
	if req.Name != nil {
		distributor.Name = *req.Name
	}
	if req.ContactEmail != nil {
		distributor.ContactEmail = req.ContactEmail
	}
	if req.ContactPhone != nil {
		distributor.ContactPhone = req.ContactPhone
	}
	if req.Address != nil {
		distributor.Address = req.Address
	}
	if req.LicenseNumber != nil {
		distributor.LicenseNumber = req.LicenseNumber
	}

	if err := h.distributorService.Update(ctx, tenantID, distributor); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, distributor)
}

// DeleteDistributor handles deleting a distributor
func (h *DistributorHandlers) DeleteDistributor(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("distributors:delete")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	distributorIDStr := c.Param("id")
	if distributorIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Distributor ID is required")
	}

	distributorID, err := uuid.Parse(distributorIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid distributor ID format")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Check if distributor exists before deleting
	_, err = h.distributorService.GetByID(ctx, tenantID, distributorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Distributor not found")
	}

	if err := h.distributorService.Delete(ctx, tenantID, distributorID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete distributor")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Distributor deleted successfully",
	})
}