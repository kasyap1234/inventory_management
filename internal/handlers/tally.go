package handlers

import (
	"net/http"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/jobs"
	"agromart2/internal/models"

	"github.com/labstack/echo/v4"

	"github.com/hibiken/asynq"
)

// TallyHandlers handles tally-related API endpoints
type TallyHandlers struct {
	tallyExporter *jobs.TallyExporter
	tallyImporter *jobs.TallyImporter
	asynqClient   *asynq.Client
}

// NewTallyHandlers creates a new TallyHandlers instance
func NewTallyHandlers(tallyExporter *jobs.TallyExporter, tallyImporter *jobs.TallyImporter, asynqClient *asynq.Client) *TallyHandlers {
	return &TallyHandlers{
		tallyExporter: tallyExporter,
		tallyImporter: tallyImporter,
		asynqClient:   asynqClient,
	}
}

// ExportTallyData handles POST /api/tally/export (non-blocking with Asynq)
func (h *TallyHandlers) ExportTallyData(c echo.Context) error {
	ctx := c.Request().Context()

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Bind request data
	var req models.ExportRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request data")
	}

	// Set default values
	if req.StartDate == "" {
		req.StartDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if req.EndDate == "" {
		req.EndDate = time.Now().Format("2006-01-02")
	}
	if req.Format == "" {
		req.Format = "csv"
	}

	// Create task
	task, err := jobs.NewTallyExportTask(tenantID, req.StartDate, req.EndDate, req.Format, req.DataType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create export task")
	}

	// Enqueue the task
	info, err := h.asynqClient.Enqueue(task)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to enqueue export task")
	}

	// Return queued response
	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"message": "Export job queued successfully",
		"job_id":  info.ID,
		"type":    "tally_export",
	})
}
// ImportTallyData handles POST /api/tally/import (non-blocking with Asynq)
func (h *TallyHandlers) ImportTallyData(c echo.Context) error {
	ctx := c.Request().Context()

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Bind request data
	var req models.ImportRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request data")
	}

	if req.DataType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "data_type is required (orders or invoices)")
	}

	if req.Data == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "data is required (CSV content)")
	}

	// Create job import request
	jobReq := jobs.ImportRequest{
		TenantID: tenantID,
		Data:     req.Data,
		DataType: req.DataType,
	}

	// Import data
	jobResult, err := h.tallyImporter.ImportData(ctx, jobReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to import data")
	}

	result := &models.ImportResult{
		RecordsProcessed: jobResult.RecordsProcessed,
		RecordsImported:  jobResult.RecordsImported,
		Errors:           jobResult.Errors,
		Message:          "Import completed",
	}

	if len(jobResult.Errors) > 0 {
		result.Message = "Import completed with errors"
	} else {
		result.Message = "Import completed successfully"
	}

	return c.JSON(http.StatusOK, result)
}