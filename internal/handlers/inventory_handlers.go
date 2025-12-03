package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"agromart2/internal/services"
	"agromart2/internal/validation"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// InventoryHandlers handles inventory-related HTTP requests
type InventoryHandlers struct {
	inventoryService services.InventoryService
	auditLogsService services.AuditLogsService
	rbacMiddleware   *middleware.RBACMiddleware
}

// NewInventoryHandlers creates a new inventory handlers instance
func NewInventoryHandlers(inventoryService services.InventoryService, auditLogsService services.AuditLogsService, rbacMiddleware *middleware.RBACMiddleware) *InventoryHandlers {
	return &InventoryHandlers{
		inventoryService: inventoryService,
		auditLogsService: auditLogsService,
		rbacMiddleware:   rbacMiddleware,
	}
}

// ListInventoriesRequest represents query parameters for listing inventories
type ListInventoriesRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

// ListInventories handles getting a list of inventories with tenant filtering
// Note: Permission check is handled by route middleware - no duplicate check needed here
func (h *InventoryHandlers) ListInventories(c echo.Context) error {
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
	tenantID, ok := common.GetTenantIDFromContext(ctx)
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
// Note: Permission check is handled by route middleware - no duplicate check needed here
func (h *InventoryHandlers) CreateInventory(c echo.Context) error {
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
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Create new inventory
	inventory := &models.Inventory{
		ID:          uuid.New(),
		TenantID:    tenantID,
		WarehouseID: req.WarehouseID,
		ProductID:   req.ProductID,
		Quantity:    req.Quantity,
	}

	if err := h.inventoryService.Create(ctx, inventory); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, inventory)
}

// GetInventory handles getting inventory details by ID
// Note: Permission check is handled by route middleware - no duplicate check needed here
func (h *InventoryHandlers) GetInventory(c echo.Context) error {
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
	tenantID, ok := common.GetTenantIDFromContext(ctx)
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
// Note: Permission check is handled by route middleware - no duplicate check needed here
func (h *InventoryHandlers) UpdateInventory(c echo.Context) error {
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
	tenantID, ok := common.GetTenantIDFromContext(ctx)
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

	if err := h.inventoryService.Update(ctx, inventory); err != nil {
		// Handle unique constraint violation
		if err.Error() == "UNIQUE constraint failed" || err.Error() == "pq: duplicate key value violates unique constraint" {
			return echo.NewHTTPError(http.StatusConflict, "Inventory record already exists for this warehouse and product combination")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, inventory)
}

// DeleteInventory handles deleting an inventory record
// Note: Permission check is handled by route middleware - no duplicate check needed here
func (h *InventoryHandlers) DeleteInventory(c echo.Context) error {
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
	tenantID, ok := common.GetTenantIDFromContext(ctx)
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
	ProductID  uuid.UUID `json:"product_id" validate:"required"`
	Adjustment int       `json:"adjustment" validate:"required"`
	Reason     string    `json:"reason" validate:"required"`
}

// AdjustStock handles stock adjustments
// Note: Permission check is handled by route middleware - no duplicate check needed here
func (h *InventoryHandlers) AdjustStock(c echo.Context) error {
	ctx := c.Request().Context()

	var req AdjustStockRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if strings.TrimSpace(req.Reason) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Adjustment reason is required")
	}

	if req.Adjustment == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Adjustment must be non-zero")
	}

	req.Reason = validation.SanitizeHTMLElement(strings.TrimSpace(req.Reason))

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	userID, _ := common.GetUserIDFromContext(ctx)

	if err := h.inventoryService.AdjustStock(ctx, tenantID, req.ProductID, req.Adjustment, req.Reason, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Stock adjusted successfully",
	})
}

// GetInventoryHistory returns audit history for an inventory record
// Note: Permission check is handled by route middleware - no duplicate check needed here
func (h *InventoryHandlers) GetInventoryHistory(c echo.Context) error {
	ctx := c.Request().Context()

	inventoryIDStr := c.Param("id")
	inventoryID, err := common.ValidateUUID(inventoryIDStr, "inventory_id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	if _, err := h.inventoryService.GetByID(ctx, tenantID, inventoryID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Inventory not found")
	}

	limit := 20
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if parsed, parseErr := strconv.Atoi(limitParam); parseErr == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if parsed, parseErr := strconv.Atoi(offsetParam); parseErr == nil && parsed >= 0 {
			offset = parsed
		}
	}

	history, err := h.auditLogsService.GetEntityHistory(ctx, tenantID, "inventory", inventoryID.String(), limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load inventory history")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"history": history,
		"limit":   limit,
		"offset":  offset,
	})
}

// CheckAvailabilityRequest represents availability check request
type CheckAvailabilityRequest struct {
	WarehouseID uuid.UUID `json:"warehouse_id" validate:"required"`
	ProductID   uuid.UUID `json:"product_id" validate:"required"`
	Quantity    int       `json:"quantity" validate:"required,min=1"`
}

// CheckAvailability handles stock availability queries
// Note: Permission check is handled by route middleware - no duplicate check needed here
func (h *InventoryHandlers) CheckAvailability(c echo.Context) error {
	ctx := c.Request().Context()

	var req CheckAvailabilityRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if req.Quantity < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Quantity must be positive")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get inventory for the warehouse and product
	inventory, err := h.inventoryService.GetByWarehouseAndProduct(ctx, tenantID, req.WarehouseID, req.ProductID)
	if err != nil {
		// If not found, assume no stock
		return c.JSON(http.StatusOK, map[string]interface{}{
			"available":          false,
			"requested":          req.Quantity,
			"available_quantity": 0,
		})
	}

	available := inventory.Quantity >= req.Quantity
	return c.JSON(http.StatusOK, map[string]interface{}{
		"available":          available,
		"requested":          req.Quantity,
		"available_quantity": inventory.Quantity,
	})
}

// TransferStock handles stock transfers between warehouses
// Note: Permission check is handled by route middleware - no duplicate check needed here
func (h *InventoryHandlers) TransferStock(c echo.Context) error {
	ctx := c.Request().Context()

	type TransferRequest struct {
		ProductID       uuid.UUID `json:"product_id" validate:"required"`
		FromWarehouseID uuid.UUID `json:"from_warehouse_id" validate:"required"`
		ToWarehouseID   uuid.UUID `json:"to_warehouse_id" validate:"required"`
		Quantity        int       `json:"quantity" validate:"required,min=1"`
	}

	var req TransferRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if req.Quantity < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Quantity must be positive")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
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
// Note: Permission check is handled by route middleware - no duplicate check needed here
func (h *InventoryHandlers) SearchInventories(c echo.Context) error {
	ctx := c.Request().Context()
	start := time.Now()

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var filter models.InventorySearchFilter
	if err := c.Bind(&filter); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid search parameters")
	}

	inventories, err := h.inventoryService.AdvancedSearch(ctx, tenantID, &filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search inventories")
	}

	// Record search usage (non-blocking)
	filterCount := 0
	if strings.TrimSpace(filter.Query) != "" { /* not a filter count */
	}
	if filter.WarehouseID != nil {
		filterCount++
	}
	if filter.ProductID != nil {
		filterCount++
	}
	duration := time.Since(start)
	common.PublishEvent(ctx, "search_performed", map[string]interface{}{
		"entity_type":      "inventory",
		"search_term":      strings.TrimSpace(filter.Query),
		"filter_count":     filterCount,
		"result_count":     len(inventories),
		"response_time_ms": duration.Milliseconds(),
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"inventories": inventories,
	})
}

// BulkAdjustStockRequest represents bulk stock adjustment request
type BulkAdjustStockRequest struct {
	Adjustments []struct {
		ProductID  uuid.UUID `json:"product_id" validate:"required"`
		Adjustment int       `json:"adjustment" validate:"required"`
		Reason     string    `json:"reason" validate:"required"`
	} `json:"adjustments" validate:"required,min=1"`
}

// BulkAdjustStock handles multiple stock adjustments
func (h *InventoryHandlers) BulkAdjustStock(c echo.Context) error {
	ctx := c.Request().Context()

	var req BulkAdjustStockRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	userID, _ := common.GetUserIDFromContext(ctx)

	// Convert request to service model
	// We need to import repositories package to use BulkAdjustmentItem
	// But handlers package usually depends on service interface, not repositories implementation details
	// However, BulkAdjustmentItem is defined in repositories package and used in service interface
	// So we need to import agromart2/internal/repositories

	// Since I cannot easily add imports with replace_file_content if they are far away,
	// I will assume the import is added or I will add it in a separate step.
	// For now, I will use repositories.BulkAdjustmentItem assuming import exists.

	// Wait, I should check imports first.
	// I'll add the import in a separate step.

	adjustments := make([]repositories.BulkAdjustmentItem, len(req.Adjustments))
	for i, adj := range req.Adjustments {
		adjustments[i] = repositories.BulkAdjustmentItem{
			ProductID:  adj.ProductID,
			Quantity:   adj.Adjustment,
			Reason:     adj.Reason,
			AdjustedBy: userID,
		}
	}

	if err := h.inventoryService.BulkAdjustStock(ctx, tenantID, adjustments); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Bulk stock adjustment completed successfully",
	})
}

// BulkDeleteInventoryRequest represents bulk inventory deletion request
type BulkDeleteInventoryRequest struct {
	InventoryIDs []uuid.UUID `json:"inventory_ids" validate:"required,min=1"`
}

// BulkDeleteInventory handles deleting multiple inventory records
func (h *InventoryHandlers) BulkDeleteInventory(c echo.Context) error {
	ctx := c.Request().Context()

	var req BulkDeleteInventoryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	if err := h.inventoryService.BulkDelete(ctx, tenantID, req.InventoryIDs); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to bulk delete inventory")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Bulk inventory deletion completed successfully",
	})
}
