package handlers

import (
	"net/http"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type InventoryReservationHandlers struct {
	reservationRepo repositories.InventoryReservationRepository
}

func NewInventoryReservationHandlers(reservationRepo repositories.InventoryReservationRepository) *InventoryReservationHandlers {
	return &InventoryReservationHandlers{
		reservationRepo: reservationRepo,
	}
}

// CreateReservation handles POST /inventory-reservations
func (h *InventoryReservationHandlers) CreateReservation(c echo.Context) error {
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
		ProductID     uuid.UUID  `json:"product_id" validate:"required"`
		WarehouseID   *uuid.UUID `json:"warehouse_id"`
		ReservationID string     `json:"reservation_id" validate:"required"`
		Quantity      int        `json:"quantity" validate:"required,gt=0"`
		ExpiresAt     *time.Time `json:"expires_at"`
		OrderID       *uuid.UUID `json:"order_id"`
		Notes         string     `json:"notes"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Convert pointer UUIDs to regular UUIDs
	warehouseID := uuid.Nil
	if req.WarehouseID != nil {
		warehouseID = *req.WarehouseID
	}

	orderID := uuid.Nil
	if req.OrderID != nil {
		orderID = *req.OrderID
	}

	reservation := &repositories.InventoryReservation{
		TenantID:      tenantID,
		ProductID:     req.ProductID,
		WarehouseID:   warehouseID,
		ReservationID: req.ReservationID,
		Quantity:      req.Quantity,
		ReservedBy:    userID,
		ExpiresAt:     req.ExpiresAt,
		Status:        "active",
		OrderID:       orderID,
		Notes:         req.Notes,
	}

	if err := h.reservationRepo.Create(ctx, reservation); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, reservation)
}

// ListReservations handles GET /inventory-reservations
func (h *InventoryReservationHandlers) ListReservations(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Parse query parameters
	status := c.QueryParam("status")
	productIDStr := c.QueryParam("product_id")

	var reservations []*repositories.InventoryReservation
	var err error

	// Apply filters
	if status != "" {
		reservations, err = h.reservationRepo.GetByStatus(ctx, tenantID, status)
	} else if productIDStr != "" {
		productID, parseErr := uuid.Parse(productIDStr)
		if parseErr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
		}
		reservations, err = h.reservationRepo.GetByProduct(ctx, tenantID, productID)
	} else {
		// Default: list all with pagination (if it exists in the interface)
		// For now, let's get by status "active"
		reservations, err = h.reservationRepo.GetByStatus(ctx, tenantID, "active")
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"reservations": reservations,
		"count":        len(reservations),
	})
}

// GetReservation handles GET /inventory-reservations/:id
func (h *InventoryReservationHandlers) GetReservation(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid reservation ID")
	}

	reservation, err := h.reservationRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Reservation not found")
	}

	return c.JSON(http.StatusOK, reservation)
}

// UpdateReservationStatus handles PUT /inventory-reservations/:id/status
func (h *InventoryReservationHandlers) UpdateReservationStatus(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid reservation ID")
	}

	var req struct {
		Status string `json:"status" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validate status
	validStatuses := map[string]bool{
		"active": true, "expired": true, "released": true, "committed": true,
	}
	if !validStatuses[req.Status] {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid status")
	}

	if err := h.reservationRepo.UpdateStatus(ctx, tenantID, id, req.Status); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Status updated successfully",
	})
}

// DeleteReservation handles DELETE /inventory-reservations/:id
func (h *InventoryReservationHandlers) DeleteReservation(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid reservation ID")
	}

	if err := h.reservationRepo.Delete(ctx, tenantID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Reservation deleted successfully",
	})
}
