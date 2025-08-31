package handlers

import (
	"log"
	"net/http"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// WarehouseHandlers handles warehouse-related HTTP requests
type WarehouseHandlers struct {
	warehouseService services.WarehouseService
	rbacMiddleware   *middleware.RBACMiddleware
}

// NewWarehouseHandlers creates a new warehouse handlers instance
func NewWarehouseHandlers(warehouseService services.WarehouseService, rbacMiddleware *middleware.RBACMiddleware) *WarehouseHandlers {
	return &WarehouseHandlers{
		warehouseService: warehouseService,
		rbacMiddleware:   rbacMiddleware,
	}
}

// ListWarehousesRequest represents query parameters for listing warehouses
type ListWarehousesRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

// ListWarehouses handles getting a list of warehouses with tenant filtering
func (h *WarehouseHandlers) ListWarehouses(c echo.Context) error {
	log.Printf("DEBUG: ListWarehouses handler called")

	// TODO: Enable RBAC for warehouses once permissions are configured
	// err := h.rbacMiddleware.RequirePermission("warehouses:list")(func(c echo.Context) error {
	// 	return nil
	// })(c)
	// if err != nil {
	// 	log.Printf("DEBUG: ListWarehouses RBAC failed: %v", err)
	// 	return err // RBAC middleware will return appropriate error
	// }

	log.Printf("DEBUG: ListWarehouses handlers (RBAC disabled)")

	ctx := c.Request().Context()
	log.Printf("DEBUG: ListWarehouses context retrieved")

	var req ListWarehousesRequest
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
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get warehouses from the tenant
	warehouses, err := h.warehouseService.List(ctx, tenantID, req.Limit, req.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list warehouses")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"warehouses": warehouses,
		"limit":      req.Limit,
		"offset":     req.Offset,
	})
}

// CreateWarehouseRequest represents the warehouse creation request payload
type CreateWarehouseRequest struct {
	Name          string  `json:"name" validate:"required"`
	Address       *string `json:"address"`
	Capacity      *int    `json:"capacity" validate:"required"`
	LicenseNumber *string `json:"license_number"`
}

// CreateWarehouse handles creating a new warehouse
func (h *WarehouseHandlers) CreateWarehouse(c echo.Context) error {
	// TODO: Enable RBAC for warehouses once permissions are configured
	// err := h.rbacMiddleware.RequirePermission("warehouses:create")(func(c echo.Context) error {
	// 	return nil
	// })(c)
	// if err != nil {
	// 	return err
	// }

	ctx := c.Request().Context()

	var req CreateWarehouseRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Name is required")
	}
	if req.Capacity == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Capacity is required")
	}

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Create new warehouse
	warehouse := &models.Warehouse{
		Name:          req.Name,
		Address:       req.Address,
		Capacity:      req.Capacity,
		LicenseNumber: req.LicenseNumber,
	}

	if err := h.warehouseService.Create(ctx, tenantID, warehouse); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, warehouse)
}

// GetWarehouse handles getting warehouse details by ID
func (h *WarehouseHandlers) GetWarehouse(c echo.Context) error {
	// TODO: Enable RBAC for warehouses once permissions are configured
	// err := h.rbacMiddleware.RequirePermission("warehouses:read")(func(c echo.Context) error {
	// 	return nil
	// })(c)
	// if err != nil {
	// 	return err
	// }

	ctx := c.Request().Context()

	warehouseIDStr := c.Param("id")
	if warehouseIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Warehouse ID is required")
	}

	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid warehouse ID format")
	}

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	log.Printf("DEBUG: ListWarehouses tenant ID: %s, ok: %v", tenantID.String(), ok)
	if !ok {
		log.Printf("DEBUG: ListWarehouses - tenant not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get warehouse details
	warehouse, err := h.warehouseService.GetByID(ctx, tenantID, warehouseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Warehouse not found")
	}

	return c.JSON(http.StatusOK, warehouse)
}

// UpdateWarehouseRequest represents the warehouse update request payload
type UpdateWarehouseRequest struct {
	Name          *string `json:"name"`
	Address       *string `json:"address"`
	Capacity      *int    `json:"capacity"`
	LicenseNumber *string `json:"license_number"`
}

// UpdateWarehouse handles updating warehouse details
func (h *WarehouseHandlers) UpdateWarehouse(c echo.Context) error {
	// TODO: Enable RBAC for warehouses once permissions are configured
	// err := h.rbacMiddleware.RequirePermission("warehouses:update")(func(c echo.Context) error {
	// 	return nil
	// })(c)
	// if err != nil {
	// 	return err
	// }

	ctx := c.Request().Context()

	warehouseIDStr := c.Param("id")
	if warehouseIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Warehouse ID is required")
	}

	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid warehouse ID format")
	}

	var req UpdateWarehouseRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get existing warehouse
	warehouse, err := h.warehouseService.GetByID(ctx, tenantID, warehouseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Warehouse not found")
	}

	// Update fields if provided
	if req.Name != nil {
		warehouse.Name = *req.Name
	}
	if req.Address != nil {
		warehouse.Address = req.Address
	}
	if req.Capacity != nil {
		warehouse.Capacity = req.Capacity
	}
	if req.LicenseNumber != nil {
		warehouse.LicenseNumber = req.LicenseNumber
	}

	if err := h.warehouseService.Update(ctx, tenantID, warehouse); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, warehouse)
}

// DeleteWarehouse handles deleting a warehouse
func (h *WarehouseHandlers) DeleteWarehouse(c echo.Context) error {
	// TODO: Enable RBAC for warehouses once permissions are configured
	// err := h.rbacMiddleware.RequirePermission("warehouses:delete")(func(c echo.Context) error {
	// 	return nil
	// })(c)
	// if err != nil {
	// 	return err
	// }

	ctx := c.Request().Context()

	warehouseIDStr := c.Param("id")
	if warehouseIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Warehouse ID is required")
	}

	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid warehouse ID format")
	}

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Check if warehouse exists before deleting
	_, err = h.warehouseService.GetByID(ctx, tenantID, warehouseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Warehouse not found")
	}

	if err := h.warehouseService.Delete(ctx, tenantID, warehouseID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete warehouse")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Warehouse deleted successfully",
	})
}