package handlers

import (
	"net/http"
	"strconv"
	"time"

	"agromart2/internal/analytics"
	"agromart2/internal/jobs"
	"agromart2/internal/middleware"
	"agromart2/internal/repositories"

	"github.com/labstack/echo/v4"
)

type JobHandlers struct {
	tallyExporter      *jobs.TallyExporter
	tallyImporter      *jobs.TallyImporter
	inventoryAlerts    *jobs.InventoryAlertService
	analyticsRefresh   *jobs.AnalyticsRefreshService
	orderRepo          repositories.OrderRepository
	invoiceRepo        repositories.InvoiceRepository
	productRepo        repositories.ProductRepository
	inventoryRepo      repositories.InventoryRepository
}

func NewJobHandlers(
	tallyExporter *jobs.TallyExporter,
	tallyImporter *jobs.TallyImporter,
	inventoryAlerts *jobs.InventoryAlertService,
	analyticsRefresh *jobs.AnalyticsRefreshService,
	orderRepo repositories.OrderRepository,
	invoiceRepo repositories.InvoiceRepository,
	productRepo repositories.ProductRepository,
	inventoryRepo repositories.InventoryRepository,
) *JobHandlers {
	return &JobHandlers{
		tallyExporter:    tallyExporter,
		tallyImporter:    tallyImporter,
		inventoryAlerts:  inventoryAlerts,
		analyticsRefresh: analyticsRefresh,
		orderRepo:        orderRepo,
		invoiceRepo:      invoiceRepo,
		productRepo:      productRepo,
		inventoryRepo:    inventoryRepo,
	}
}

// ExportInvoices handler
func (h *JobHandlers) ExportInvoices(c echo.Context) error {
	ctx := c.Request().Context()

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	req := &jobs.ExportRequest{
		TenantID: tenantID,
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	analyticsService := analytics.NewAnalyticsService(h.orderRepo, h.invoiceRepo, h.inventoryRepo, h.productRepo, nil)
	data, err := analyticsService.CalculateTenantAnalytics(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get analytics data")
	}

	return c.JSON(http.StatusOK, data)
}

