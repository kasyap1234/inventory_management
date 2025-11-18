package handlers

import (
	"net/http"

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

	if err := h.batchService.CreateBatch(c.Request().Context(), batch); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, batch)
}

// GetBatchesByProduct handles fetching all batches for a product
func (h *BatchHandler) GetBatchesByProduct(c echo.Context) error {
	productIDStr := c.Param("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	batches, err := h.batchService.GetBatchesByProduct(c.Request().Context(), productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, batches)
}

// UpdateBatch handles updating a batch
func (h *BatchHandler) UpdateBatch(c echo.Context) error {
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

	if err := h.batchService.UpdateBatch(c.Request().Context(), batch); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, batch)
}

// GetBatch handles fetching a single batch
func (h *BatchHandler) GetBatch(c echo.Context) error {
	batchIDStr := c.Param("id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid batch ID")
	}

	batch, err := h.batchService.GetBatch(c.Request().Context(), batchID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if batch == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Batch not found")
	}

	return c.JSON(http.StatusOK, batch)
}
