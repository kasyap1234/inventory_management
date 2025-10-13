package handlers

import (
	"net/http"
	"strconv"
	"time"

	"agromart2/internal/analytics"
	"agromart2/internal/common"
	"agromart2/internal/jobs"
	"agromart2/internal/repositories"

	"github.com/labstack/echo/v4"
)

type JobHandlers struct {
	tallyExporter    *jobs.TallyExporter
	tallyImporter    *jobs.TallyImporter
	inventoryAlerts  *jobs.InventoryAlertService
	analyticsRefresh *jobs.AnalyticsRefreshService
	analyticsService *analytics.AnalyticsService
	orderRepo        repositories.OrderRepository
	invoiceRepo      repositories.InvoiceRepository
	productRepo      repositories.ProductRepository
	inventoryRepo    repositories.InventoryRepository
	inspector        jobs.JobInspector
}

func NewJobHandlers(
	tallyExporter *jobs.TallyExporter,
	tallyImporter *jobs.TallyImporter,
	inventoryAlerts *jobs.InventoryAlertService,
	analyticsRefresh *jobs.AnalyticsRefreshService,
	analyticsService *analytics.AnalyticsService,
	orderRepo repositories.OrderRepository,
	invoiceRepo repositories.InvoiceRepository,
	productRepo repositories.ProductRepository,
	inventoryRepo repositories.InventoryRepository,
	inspector jobs.JobInspector,
) *JobHandlers {
	return &JobHandlers{
		tallyExporter:    tallyExporter,
		tallyImporter:    tallyImporter,
		inventoryAlerts:  inventoryAlerts,
		analyticsRefresh: analyticsRefresh,
		analyticsService: analyticsService,
		orderRepo:        orderRepo,
		invoiceRepo:      invoiceRepo,
		productRepo:      productRepo,
		inventoryRepo:    inventoryRepo,
		inspector:        inspector,
	}
}

// ExportInvoices handler
func (h *JobHandlers) ExportInvoices(c echo.Context) error {
	ctx := c.Request().Context()

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get request data
	req := &jobs.ExportRequest{
		TenantID: tenantID,
	}

	startDate := c.QueryParam("start_date")
	if startDate == "" {
		// Default to last 30 days
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	req.StartDate = startDate

	endDate := c.QueryParam("end_date")
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}
	req.EndDate = endDate

	req.Format = c.QueryParam("format")
	if req.Format == "" {
		req.Format = "csv"
	}

	result, err := h.tallyExporter.ExportInvoicesForTenant(ctx, *req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to export invoices")
	}

	// Set headers for file download
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+result.FileName)
	c.Response().Header().Set("Content-Type", "text/csv")

	return c.String(http.StatusOK, result.FileContent)
}

// ExportOrders handler
func (h *JobHandlers) ExportOrders(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	req := &jobs.ExportRequest{
		TenantID:  tenantID,
		StartDate: c.QueryParam("start_date"),
		EndDate:   c.QueryParam("end_date"),
		Format:    c.QueryParam("format"),
	}

	if req.StartDate == "" {
		req.StartDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if req.EndDate == "" {
		req.EndDate = time.Now().Format("2006-01-02")
	}
	if req.Format == "" {
		req.Format = "csv"
	}

	result, err := h.tallyExporter.ExportOrdersForTenant(ctx, *req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to export orders")
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+result.FileName)
	c.Response().Header().Set("Content-Type", "text/csv")

	return c.String(http.StatusOK, result.FileContent)
}

// ImportData handler
func (h *JobHandlers) ImportData(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req jobs.ImportRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request data")
	}

	req.TenantID = tenantID

	if req.DataType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "data_type is required (orders or invoices)")
	}

	result, err := h.tallyImporter.ImportData(ctx, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to import data")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Import completed",
		"result":  result,
	})
}

// GetInventoryAlerts handler
func (h *JobHandlers) GetInventoryAlerts(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	threshold := 10 // Default threshold
	if thresholdParam := c.QueryParam("threshold"); thresholdParam != "" {
		if t, err := strconv.Atoi(thresholdParam); err == nil {
			threshold = t
		}
	}

	alerts, err := h.inventoryAlerts.CheckLowStock(ctx, tenantID, threshold)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check inventory alerts")
	}

	// Log alerts
	h.inventoryAlerts.LogLowStockAlerts(ctx, alerts)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"alerts": alerts,
	})
}

// TriggerAnalyticsRefresh handler
func (h *JobHandlers) TriggerAnalyticsRefresh(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	err := h.analyticsRefresh.RefreshAnalyticsForTenant(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to refresh analytics")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Analytics refresh completed successfully",
	})
}

func (h *JobHandlers) GetAnalyticsData(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	data, err := h.analyticsService.CalculateTenantAnalytics(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get analytics data")
	}

	return c.JSON(http.StatusOK, data)
}

// ListJobs returns list of jobs from all queues
func (h *JobHandlers) ListJobs(c echo.Context) error {
	ctx := c.Request().Context()

	if h.inspector == nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"jobs":  []interface{}{},
			"stats": map[string]int{"total": 0, "pending": 0, "active": 0, "completed": 0, "failed": 0},
		})
	}

	// Get jobs from all states
	allJobs, stats, err := h.inspector.ListAllJobs(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list jobs")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"jobs":  allJobs,
		"stats": stats,
	})
}

// GetJob returns details for a specific job
func (h *JobHandlers) GetJob(c echo.Context) error {
	ctx := c.Request().Context()
	jobID := c.Param("id")

	if h.inspector == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Job inspector not available")
	}

	job, err := h.inspector.GetJobInfo(ctx, jobID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Job not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"job": job,
	})
}

// RetryJob retries a failed job
func (h *JobHandlers) RetryJob(c echo.Context) error {
	ctx := c.Request().Context()
	jobID := c.Param("id")

	if h.inspector == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Job inspector not available")
	}

	if err := h.inspector.RetryJob(ctx, jobID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retry job")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Job queued for retry",
		"job_id":  jobID,
	})
}

// CancelJob cancels a pending or active job
func (h *JobHandlers) CancelJob(c echo.Context) error {
	ctx := c.Request().Context()
	jobID := c.Param("id")

	if h.inspector == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Job inspector not available")
	}

	if err := h.inspector.CancelJob(ctx, jobID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to cancel job")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Job cancelled successfully",
		"job_id":  jobID,
	})
}

// GetJobStats returns job queue statistics
func (h *JobHandlers) GetJobStats(c echo.Context) error {
	ctx := c.Request().Context()

	if h.inspector == nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"stats": map[string]int{"total": 0, "pending": 0, "active": 0, "completed": 0, "failed": 0},
		})
	}

	stats, err := h.inspector.GetQueueStats(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get job statistics")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"stats": stats,
	})
}
