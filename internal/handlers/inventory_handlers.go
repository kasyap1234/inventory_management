package handlers

import (
	"net/http"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// InventoryHandlers handles inventory-related HTTP requests
type InventoryHandlers struct {
	inventoryService services.InventoryService
	rbacMiddleware   *middleware.RBACMiddleware
}

// NewInventoryHandlers creates a new inventory handlers instance
func NewInventoryHandlers(inventoryService services.InventoryService, rbacMiddleware *middleware.RBACMiddleware) *InventoryHandlers {
	return &InventoryHandlers{
		inventoryService: inventoryService,
		rbacMiddleware:   rbacMiddleware,
	}
}

// ListInventoriesRequest represents query parameters for listing inventories
type ListInventoriesRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

// ListInventories handles getting a list of inventories with tenant filtering
func (h *InventoryHandlers) ListInventories(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("inventories:list")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err // RBAC middleware will return appropriate error
	}

	ctx := c.Request().Context()

	var req ListInventoriesRequest
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

	// Get inventories from the tenant
	inventories, err := h.inventoryService.List(ctx, tenantID, req.Limit, req.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list inventories")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"inventories": inventories,
		"limit":       req.Limit,
		"offset":      req.Offset,
	})
}

// CreateInventoryRequest represents the inventory creation request payload
type CreateInventoryRequest struct {
	WarehouseID uuid.UUID `json:"warehouse_id" validate:"required"`
	ProductID   uuid.UUID `json:"product_id" validate:"required"`
	Quantity    int       `json:"quantity" validate:"required,min=0"`
}

// CreateInventory handles creating/updating inventory records (handles unique constraint)
func (h *InventoryHandlers) CreateInventory(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("inventories:create")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	var req CreateInventoryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if req.Quantity < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Quantity cannot be negative")
	}

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Create new inventory
	inventory := &models.Inventory{
		WarehouseID: req.WarehouseID,
		ProductID:   req.ProductID,
		Quantity:    req.Quantity,
	}

	if err := h.inventoryService.Create(ctx, tenantID, inventory); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, inventory)
}

// GetInventory handles getting inventory details by ID
func (h *InventoryHandlers) GetInventory(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("inventories:read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	inventoryIDStr := c.Param("id")
	if inventoryIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Inventory ID is required")
	}

	inventoryID, err := uuid.Parse(inventoryIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid inventory ID format")
	}

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get inventory details
	inventory, err := h.inventoryService.GetByID(ctx, tenantID, inventoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Inventory not found")
	}

	return c.JSON(http.StatusOK, inventory)
}

// UpdateInventoryRequest represents the inventory update request payload
type UpdateInventoryRequest struct {
	WarehouseID *uuid.UUID `json:"warehouse_id"`
	ProductID   *uuid.UUID `json:"product_id"`
	Quantity    *int       `json:"quantity"`
}

// UpdateInventory handles updating inventory details
func (h *InventoryHandlers) UpdateInventory(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("inventories:update")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	inventoryIDStr := c.Param("id")
	if inventoryIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Inventory ID is required")
	}

	inventoryID, err := uuid.Parse(inventoryIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid inventory ID format")
	}

	var req UpdateInventoryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validate quantity if provided
	if req.Quantity != nil && *req.Quantity < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Quantity cannot be negative")
	}

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get existing inventory
	inventory, err := h.inventoryService.GetByID(ctx, tenantID, inventoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Inventory not found")
	}

	// Update fields if provided
	if req.WarehouseID != nil {
		inventory.WarehouseID = *req.WarehouseID
	}
	if req.ProductID != nil {
		inventory.ProductID = *req.ProductID
	}
	if req.Quantity != nil {
		inventory.Quantity = *req.Quantity
	}

	if err := h.inventoryService.Update(ctx, tenantID, inventory); err != nil {
		// Handle unique constraint violation
		if err.Error() == "UNIQUE constraint failed" || err.Error() == "pq: duplicate key value violates unique constraint" {
			return echo.NewHTTPError(http.StatusConflict, "Inventory record already exists for this warehouse and product combination")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, inventory)
}

// DeleteInventory handles deleting an inventory record
func (h *InventoryHandlers) DeleteInventory(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("inventories:delete")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	inventoryIDStr := c.Param("id")
	if inventoryIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Inventory ID is required")
	}

	inventoryID, err := uuid.Parse(inventoryIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid inventory ID format")
	}

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Check if inventory exists before deleting
	_, err = h.inventoryService.GetByID(ctx, tenantID, inventoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Inventory not found")
	}

	if err := h.inventoryService.Delete(ctx, tenantID, inventoryID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete inventory")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Inventory deleted successfully",
	})
}

// AdjustStockRequest represents stock adjustment request
type AdjustStockRequest struct {
	WarehouseID     uuid.UUID `json:"warehouse_id" validate:"required"`
	ProductID       uuid.UUID `json:"product_id" validate:"required"`
	QuantityChange  int       `json:"quantity_change" validate:"required"`
}

// AdjustStock handles stock adjustments
func (h *InventoryHandlers) AdjustStock(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("inventories:update")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	var req AdjustStockRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	if err := h.inventoryService.AdjustStock(ctx, tenantID, req.WarehouseID, req.ProductID, req.QuantityChange); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Stock adjusted successfully",
	})
}

// CheckAvailabilityRequest represents availability check request
type CheckAvailabilityRequest struct {
	WarehouseID uuid.UUID `json:"warehouse_id" validate:"required"`
	ProductID   uuid.UUID `json:"product_id" validate:"required"`
	Quantity    int       `json:"quantity" validate:"required,min=1"`
}

// CheckAvailability handles stock availability queries
func (h *InventoryHandlers) CheckAvailability(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("inventories:read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	var req CheckAvailabilityRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if req.Quantity < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Quantity must be positive")
	}

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get inventory for the warehouse and product
	inventory, err := h.inventoryService.GetByWarehouseAndProduct(ctx, tenantID, req.WarehouseID, req.ProductID)
	if err != nil {
		// If not found, assume no stock
		return c.JSON(http.StatusOK, map[string]interface{}{
			"available": false,
			"requested": req.Quantity,
			"available_quantity": 0,
		})
	}

	available := inventory.Quantity >= req.Quantity
	return c.JSON(http.StatusOK, map[string]interface{}{
		"available": available,
		"requested": req.Quantity,
		"available_quantity": inventory.Quantity,
	})
}

// TransferStock handles stock transfers between warehouses
func (h *InventoryHandlers) TransferStock(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("inventories:update")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	type TransferRequest struct {
		ProductID        uuid.UUID `json:"product_id" validate:"required"`
		FromWarehouseID  uuid.UUID `json:"from_warehouse_id" validate:"required"`
		ToWarehouseID    uuid.UUID `json:"to_warehouse_id" validate:"required"`
		Quantity         int       `json:"quantity" validate:"required,min=1"`
	}

	var req TransferRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if req.Quantity < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Quantity must be positive")
	}

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	if err := h.inventoryService.Transfer(ctx, tenantID, req.ProductID, req.FromWarehouseID, req.ToWarehouseID, req.Quantity); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Stock transferred successfully",
	})
}

// SearchInventories handles advanced search with filters
func (h *InventoryHandlers) SearchInventories(c echo.Context) error {
	// Use RBAC middleware directly
	err := h.rbacMiddleware.RequirePermission("inventories:list")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	var filter models.InventorySearchFilter
	if err := c.Bind(&filter); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid search parameters")
	}

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	inventories, err := h.inventoryService.AdvancedSearch(ctx, tenantID, &filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search inventories")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"inventories": inventories,
	})
}