package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"agromart2/internal/analytics"
	"agromart2/internal/common"
	"agromart2/internal/jobs"
	"agromart2/internal/repositories"

	"github.com/hibiken/asynq"
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
	inspector        *asynq.Inspector
	client           *asynq.Client
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
	inspector *asynq.Inspector,
	client *asynq.Client,
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
		client:           client,
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

func (h *JobHandlers) ListJobs(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	queue := c.QueryParam("queue")
	if queue == "" {
		queue = "default"
	}

	state := c.QueryParam("state")
	if state == "" {
		state = "pending"
	}

	page := 1
	if pageParam := c.QueryParam("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var tasks []*asynq.TaskInfo
	var err error

	switch state {
	case "pending":
		tasks, err = h.inspector.ListPendingTasks(queue, asynq.PageSize(limit), asynq.Page(page-1))
	case "active", "in_progress":
		tasks, err = h.inspector.ListActiveTasks(queue, asynq.PageSize(limit), asynq.Page(page-1))
	case "scheduled":
		tasks, err = h.inspector.ListScheduledTasks(queue, asynq.PageSize(limit), asynq.Page(page-1))
	case "retry":
		tasks, err = h.inspector.ListRetryTasks(queue, asynq.PageSize(limit), asynq.Page(page-1))
	case "archived":
		tasks, err = h.inspector.ListArchivedTasks(queue, asynq.PageSize(limit), asynq.Page(page-1))
	case "completed":
		tasks, err = h.inspector.ListCompletedTasks(queue, asynq.PageSize(limit), asynq.Page(page-1))
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid state. Valid states: pending, active, scheduled, retry, archived, completed")
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to list jobs: %v", err))
	}

	jobList := make([]map[string]interface{}, 0, len(tasks))
	for _, task := range tasks {
		jobList = append(jobList, map[string]interface{}{
			"id":           task.ID,
			"type":         task.Type,
			"queue":        task.Queue,
			"state":        state,
			"max_retry":    task.MaxRetry,
			"retried":      task.Retried,
			"last_error":   task.LastErr,
			"next_process": task.NextProcessAt,
			"completed_at": task.CompletedAt,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tenant_id": tenantID,
		"queue":     queue,
		"state":     state,
		"page":      page,
		"limit":     limit,
		"jobs":      jobList,
		"count":     len(jobList),
	})
}

func (h *JobHandlers) GetJob(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	jobID := c.Param("id")
	if jobID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "job_id is required")
	}

	// Get queue parameter (optional, if not provided we'll search all queues)
	queue := c.QueryParam("queue")
	if queue == "" {
		queue = "default"
	}

	// Search through different states to find the job
	states := []string{"pending", "active", "scheduled", "retry", "archived", "completed"}
	
	for _, state := range states {
		var tasks []*asynq.TaskInfo
		var err error

		switch state {
		case "pending":
			tasks, err = h.inspector.ListPendingTasks(queue, asynq.PageSize(100))
		case "active":
			tasks, err = h.inspector.ListActiveTasks(queue, asynq.PageSize(100))
		case "scheduled":
			tasks, err = h.inspector.ListScheduledTasks(queue, asynq.PageSize(100))
		case "retry":
			tasks, err = h.inspector.ListRetryTasks(queue, asynq.PageSize(100))
		case "archived":
			tasks, err = h.inspector.ListArchivedTasks(queue, asynq.PageSize(100))
		case "completed":
			tasks, err = h.inspector.ListCompletedTasks(queue, asynq.PageSize(100))
		}

		if err != nil {
			continue // Skip this state if there's an error
		}

		// Search for the job in this state
		for _, task := range tasks {
			if task.ID == jobID {
				// Found the job!
				return c.JSON(http.StatusOK, map[string]interface{}{
					"tenant_id":      tenantID,
					"id":             task.ID,
					"type":           task.Type,
					"queue":          task.Queue,
					"state":          state,
					"payload":        string(task.Payload),
					"max_retry":      task.MaxRetry,
					"retried":        task.Retried,
					"last_error":     task.LastErr,
					"last_failed_at": task.LastFailedAt,
					"next_process":   task.NextProcessAt,
					"completed_at":   task.CompletedAt,
					"timeout":        task.Timeout,
					"deadline":       task.Deadline,
				})
			}
		}
	}

	// Job not found in any state
	return echo.NewHTTPError(http.StatusNotFound, map[string]interface{}{
		"message": "Job not found",
		"job_id":  jobID,
		"queue":   queue,
		"suggestion": "The job may have been deleted or the ID is incorrect. Try listing jobs with GET /jobs?queue=<queue>&state=<state>",
	})
}

func (h *JobHandlers) RetryJob(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	jobID := c.Param("id")
	if jobID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "job_id is required")
	}

	queue := c.QueryParam("queue")
	if queue == "" {
		queue = "default"
	}

	err := h.inspector.RunTask(queue, jobID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to retry job: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":   "Job retry initiated successfully",
		"tenant_id": tenantID,
		"job_id":    jobID,
		"queue":     queue,
	})
}

func (h *JobHandlers) CancelJob(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	jobID := c.Param("id")
	if jobID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "job_id is required")
	}

	queue := c.QueryParam("queue")
	if queue == "" {
		queue = "default"
	}

	err := h.inspector.DeleteTask(queue, jobID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to cancel job: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":   "Job cancelled successfully",
		"tenant_id": tenantID,
		"job_id":    jobID,
		"queue":     queue,
	})
}

func (h *JobHandlers) GetJobStats(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	queues, err := h.inspector.Queues()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get queues: %v", err))
	}

	stats := make(map[string]interface{})
	totalStats := map[string]int{
		"active":    0,
		"pending":   0,
		"scheduled": 0,
		"retry":     0,
		"archived":  0,
		"completed": 0,
	}

	queueStats := make([]map[string]interface{}, 0, len(queues))

	for _, queue := range queues {
		queueInfo, err := h.inspector.GetQueueInfo(queue)
		if err != nil {
			continue
		}

		queueData := map[string]interface{}{
			"queue":     queue,
			"active":    queueInfo.Active,
			"pending":   queueInfo.Pending,
			"scheduled": queueInfo.Scheduled,
			"retry":     queueInfo.Retry,
			"archived":  queueInfo.Archived,
			"completed": queueInfo.Completed,
			"processed": queueInfo.Processed,
			"failed":    queueInfo.Failed,
			"paused":    queueInfo.Paused,
			"size":      queueInfo.Size,
		}

		queueStats = append(queueStats, queueData)

		totalStats["active"] += queueInfo.Active
		totalStats["pending"] += queueInfo.Pending
		totalStats["scheduled"] += queueInfo.Scheduled
		totalStats["retry"] += queueInfo.Retry
		totalStats["archived"] += queueInfo.Archived
		totalStats["completed"] += queueInfo.Completed
	}

	stats["tenant_id"] = tenantID
	stats["total"] = totalStats
	stats["queues"] = queueStats
	stats["queue_count"] = len(queues)

	return c.JSON(http.StatusOK, stats)
}
