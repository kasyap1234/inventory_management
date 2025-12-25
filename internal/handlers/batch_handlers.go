package handlers

import (
	"net/http"
	"strconv"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type BatchHandler struct {
	batchService *services.BatchService
}

func NewBatchHandler(batchService *services.BatchService) *BatchHandler {
	return &BatchHandler{batchService: batchService}
}

// CreateBatch handles the creation of a new batch
func (h *BatchHandler) CreateBatch(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	productIDStr := c.Param("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	batch := new(models.Batch)
	if err := c.Bind(batch); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	batch.ProductID = productID

	if err := h.batchService.CreateBatch(ctx, tenantID, batch); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, batch)
}

// GetBatchesByProduct handles fetching all batches for a product
func (h *BatchHandler) GetBatchesByProduct(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	productIDStr := c.Param("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	batches, err := h.batchService.GetBatchesByProduct(ctx, tenantID, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, batches)
}

// UpdateBatch handles updating a batch
func (h *BatchHandler) UpdateBatch(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	batchIDStr := c.Param("id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid batch ID")
	}

	batch := new(models.Batch)
	if err := c.Bind(batch); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	batch.ID = batchID

	if err := h.batchService.UpdateBatch(ctx, tenantID, batch); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, batch)
}

// GetBatch handles fetching a single batch
func (h *BatchHandler) GetBatch(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	batchIDStr := c.Param("id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid batch ID")
	}

	batch, err := h.batchService.GetBatch(ctx, tenantID, batchID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if batch == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Batch not found")
	}

	return c.JSON(http.StatusOK, batch)
}

// DeleteBatch handles deleting a batch
func (h *BatchHandler) DeleteBatch(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	batchIDStr := c.Param("id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid batch ID")
	}

	if err := h.batchService.DeleteBatch(ctx, tenantID, batchID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// ListBatches handles fetching all batches for a tenant with pagination
func (h *BatchHandler) ListBatches(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Parse query params for pagination
	limit := 20 // default
	offset := 0
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	batches, err := h.batchService.ListBatches(ctx, tenantID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, batches)
}
