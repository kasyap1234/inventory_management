package handlers

import (
	"net/http"
	"strconv"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// StockAdjustmentHandlers handles stock adjustment HTTP requests
type StockAdjustmentHandlers struct {
	stockAdjustmentRepo repositories.StockAdjustmentRepository
	productRepo         repositories.ProductRepository
	inventoryRepo       repositories.InventoryRepository
}

// NewStockAdjustmentHandlers creates a new stock adjustment handlers instance
func NewStockAdjustmentHandlers(
	stockAdjustmentRepo repositories.StockAdjustmentRepository,
	productRepo repositories.ProductRepository,
	inventoryRepo repositories.InventoryRepository,
) *StockAdjustmentHandlers {
	return &StockAdjustmentHandlers{
		stockAdjustmentRepo: stockAdjustmentRepo,
		productRepo:         productRepo,
		inventoryRepo:       inventoryRepo,
	}
}

// CreateStockAdjustment creates a new stock adjustment
func (h *StockAdjustmentHandlers) CreateStockAdjustment(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	userID, ok := common.GetUserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	var req struct {
		ProductID      uuid.UUID `json:"product_id" validate:"required"`
		WarehouseID    uuid.UUID `json:"warehouse_id"`
		AdjustmentType string    `json:"adjustment_type" validate:"required"`
		Quantity       int       `json:"quantity" validate:"required"`
		Reason         string    `json:"reason"`
		ReferenceType  string    `json:"reference_type"`
		ReferenceID    uuid.UUID `json:"reference_id"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validate adjustment type
	validTypes := map[string]bool{
		"increase": true, "decrease": true, "reservation": true, "release": true,
		"transfer_in": true, "transfer_out": true, "correction": true,
		"damage": true, "return": true,
	}
	if !validTypes[req.AdjustmentType] {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid adjustment type")
	}

	// Get current product stock
	var previousStock int
	if req.WarehouseID != uuid.Nil {
		// Get inventory for specific warehouse
		inventory, err := h.inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, req.WarehouseID, req.ProductID)
		if err != nil {
			// If not found, assume 0 stock
			previousStock = 0
		} else {
			previousStock = inventory.Quantity
		}
	} else {
		// Get total product stock
		product, err := h.productRepo.GetByID(ctx, tenantID, req.ProductID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Product not found")
		}
		previousStock = product.Quantity
	}

	// Calculate new stock
	newStock := previousStock
	quantity := req.Quantity
	switch req.AdjustmentType {
	case "increase", "return", "transfer_in", "release":
		newStock = previousStock + quantity
	case "decrease", "damage", "transfer_out", "reservation":
		newStock = previousStock - quantity
		quantity = -quantity // Store as negative for decrease types
	case "correction":
		// For correction, quantity can be positive or negative
		newStock = previousStock + quantity
	}

	// Ensure stock doesn't go negative (except for corrections)
	if newStock < 0 && req.AdjustmentType != "correction" {
		return echo.NewHTTPError(http.StatusBadRequest, "Insufficient stock for adjustment")
	}

	// Create adjustment record
	adjustment := &repositories.StockAdjustment{
		ID:             uuid.New(),
		TenantID:       tenantID,
		ProductID:      req.ProductID,
		WarehouseID:    req.WarehouseID,
		AdjustmentType: req.AdjustmentType,
		Quantity:       quantity,
		PreviousStock:  previousStock,
		NewStock:       newStock,
		Reason:         req.Reason,
		ReferenceType:  req.ReferenceType,
		ReferenceID:    req.ReferenceID,
		AdjustedBy:     userID,
		AdjustedAt:     time.Now(),
	}

	if err := h.stockAdjustmentRepo.Create(ctx, adjustment); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Update product/inventory quantity
	if req.WarehouseID != uuid.Nil {
		// Update inventory
		inventory, err := h.inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, req.WarehouseID, req.ProductID)
		if err != nil {
			// Create new inventory record
			inventory = &models.Inventory{
				TenantID:    tenantID,
				ProductID:   req.ProductID,
				WarehouseID: req.WarehouseID,
				Quantity:    newStock,
			}
			if err := h.inventoryRepo.Create(ctx, inventory); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create inventory record")
			}
		} else {
			// Update existing inventory
			inventory.Quantity = newStock
			if err := h.inventoryRepo.Update(ctx, inventory); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update inventory")
			}
		}
	} else {
		// Update product total quantity
		product, err := h.productRepo.GetByID(ctx, tenantID, req.ProductID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Product not found")
		}
		product.Quantity = newStock
		if err := h.productRepo.Update(ctx, product); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update product quantity")
		}
	}

	return c.JSON(http.StatusCreated, adjustment)
}

// ListStockAdjustments lists stock adjustments with filters
func (h *StockAdjustmentHandlers) ListStockAdjustments(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Parse query parameters
	productIDStr := c.QueryParam("product_id")
	warehouseIDStr := c.QueryParam("warehouse_id")
	adjustmentType := c.QueryParam("adjustment_type")
	startDateStr := c.QueryParam("start_date")
	endDateStr := c.QueryParam("end_date")
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := 50
	offset := 0
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var adjustments []*repositories.StockAdjustment
	var err error

	// Apply filters
	if productIDStr != "" && warehouseIDStr != "" {
		productID, err1 := uuid.Parse(productIDStr)
		warehouseID, err2 := uuid.Parse(warehouseIDStr)
		if err1 == nil && err2 == nil {
			adjustments, err = h.stockAdjustmentRepo.GetByProductAndWarehouse(ctx, tenantID, productID, warehouseID)
		}
	} else if productIDStr != "" {
		productID, parseErr := uuid.Parse(productIDStr)
		if parseErr == nil {
			adjustments, err = h.stockAdjustmentRepo.GetByProduct(ctx, tenantID, productID)
		}
	} else if warehouseIDStr != "" {
		warehouseID, parseErr := uuid.Parse(warehouseIDStr)
		if parseErr == nil {
			adjustments, err = h.stockAdjustmentRepo.GetByWarehouse(ctx, tenantID, warehouseID)
		}
	} else if adjustmentType != "" {
		adjustments, err = h.stockAdjustmentRepo.GetByAdjustmentType(ctx, tenantID, adjustmentType)
	} else if startDateStr != "" && endDateStr != "" {
		startDate, err1 := time.Parse("2006-01-02", startDateStr)
		endDate, err2 := time.Parse("2006-01-02", endDateStr)
		if err1 == nil && err2 == nil {
			// Set end date to end of day
			endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			adjustments, err = h.stockAdjustmentRepo.GetByDateRange(ctx, tenantID, startDate, endDate)
		}
	} else {
		// Default: list all with pagination
		adjustments, err = h.stockAdjustmentRepo.List(ctx, tenantID, limit, offset)
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"adjustments": adjustments,
		"count":       len(adjustments),
		"limit":       limit,
		"offset":      offset,
	})
}

// GetStockAdjustment retrieves a specific stock adjustment
func (h *StockAdjustmentHandlers) GetStockAdjustment(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	adjustmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid adjustment ID")
	}

	adjustment, err := h.stockAdjustmentRepo.GetByID(ctx, tenantID, adjustmentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Stock adjustment not found")
	}

	return c.JSON(http.StatusOK, adjustment)
}

// GetProductStockHistory retrieves stock adjustment history for a product
func (h *StockAdjustmentHandlers) GetProductStockHistory(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	adjustments, err := h.stockAdjustmentRepo.GetByProduct(ctx, tenantID, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"adjustments": adjustments,
		"count":       len(adjustments),
	})
}
