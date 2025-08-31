package handlers

import (
	"net/http"
	"strconv"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AuditLogsHandlers handles audit logs related HTTP requests
type AuditLogsHandlers struct {
	auditLogsService services.AuditLogsService
	rbacMiddleware   *middleware.RBACMiddleware
}

// NewAuditLogsHandlers creates a new audit logs handlers instance
func NewAuditLogsHandlers(auditLogsService services.AuditLogsService, rbacMiddleware *middleware.RBACMiddleware) *AuditLogsHandlers {
	return &AuditLogsHandlers{
		auditLogsService: auditLogsService,
		rbacMiddleware:   rbacMiddleware,
	}
}

// ListAuditLogs retrieves audit logs with filtering and pagination
func (h *AuditLogsHandlers) ListAuditLogs(c echo.Context) error {
	ctx := c.Request().Context()

	// Check if user has permission to view audit logs
	err := h.rbacMiddleware.RequirePermission("audit:read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to view audit logs")
	}

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Parse query parameters
	filters := &models.AuditLogFilters{}
	if table := c.QueryParam("table"); table != "" {
		filters.TableName = &table
	}
	if recordID := c.QueryParam("record_id"); recordID != "" {
		filters.RecordID = &recordID
	}
	if action := c.QueryParam("action"); action != "" {
		filters.Action = &action
	}
	if userID := c.QueryParam("user_id"); userID != "" {
		if uid, err := uuid.Parse(userID); err == nil {
			filters.ChangedBy = &uid
		}
	}
	if startDate := c.QueryParam("start_date"); startDate != "" {
		if sd, err := time.Parse("2006-01-02T15:04:05Z", startDate); err == nil {
			filters.StartDate = &sd
		}
	}
	if endDate := c.QueryParam("end_date"); endDate != "" {
		if ed, err := time.Parse("2006-01-02T15:04:05Z", endDate); err == nil {
			filters.EndDate = &ed
		}
	}
	if includeDeleted := c.QueryParam("include_deleted"); includeDeleted == "true" {
		filters.IncludeDeleted = true
	}

	// Parse pagination
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 50 // Default limit
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	filters.Limit = limit
	filters.Offset = offset

	// Validate filters (security and performance)
	if err := h.auditLogsService.ValidateAuditFilters(filters); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get audit logs
	logs, err := h.auditLogsService.ListAuditLogs(ctx, tenantID, filters)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve audit logs")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":   logs,
		"total":  len(logs),
		"limit":  filters.Limit,
		"offset": filters.Offset,
	})
}

// GetAuditLog retrieves a specific audit log entry
func (h *AuditLogsHandlers) GetAuditLog(c echo.Context) error {
	ctx := c.Request().Context()

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Parse audit log ID from URL
	auditLogID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid audit log ID")
	}

	// Get audit log
	log, err := h.auditLogsService.GetAuditLog(ctx, tenantID, auditLogID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Audit log not found")
	}

	return c.JSON(http.StatusOK, log)
}

// GetEntityHistory retrieves audit history for a specific entity
func (h *AuditLogsHandlers) GetEntityHistory(c echo.Context) error {
	ctx := c.Request().Context()

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Parse URL parameters
	tableName := c.Param("table")
	recordID := c.Param("record_id")

	if tableName == "" || recordID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Table name and record ID are required")
	}

	// Parse pagination
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 100 // Default for entity history
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	// Get entity history
	logs, err := h.auditLogsService.GetEntityHistory(ctx, tenantID, tableName, recordID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve entity history")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":      logs,
		"total":     len(logs),
		"limit":     limit,
		"offset":    offset,
		"table":     tableName,
		"record_id": recordID,
	})
}

// GetUserActivity retrieves audit logs for a specific user's actions
func (h *AuditLogsHandlers) GetUserActivity(c echo.Context) error {
	ctx := c.Request().Context()

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Parse user ID from URL
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	// Parse pagination
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 100 // Default for user activity
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	// Get user activity logs
	logs, err := h.auditLogsService.GetUserActivity(ctx, tenantID, userID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve user activity")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":    logs,
		"total":   len(logs),
		"limit":   limit,
		"offset":  offset,
		"user_id": userID,
	})
}

// GetAuditSummary provides analytics summary for audit logs
func (h *AuditLogsHandlers) GetAuditSummary(c echo.Context) error {
	ctx := c.Request().Context()

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Parse date range
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid start_date format")
		}
	} else {
		// Default to last 30 days
		start = time.Now().AddDate(0, 0, -30)
	}

	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid end_date format")
		}
	} else {
		end = time.Now()
	}

	// Get audit summary
	summary, err := h.auditLogsService.GetAuditSummary(ctx, tenantID, start, end)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate audit summary")
	}

	return c.JSON(http.StatusOK, summary)
}

// GetTableNames returns distinct table names that have audit logs
func (h *AuditLogsHandlers) GetTableNames(c echo.Context) error {
	ctx := c.Request().Context()

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get table names
	tableNames, err := h.auditLogsService.GetTableNames(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve table names")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"table_names": tableNames,
		"count":       len(tableNames),
	})
}

// GetActions returns distinct actions that have been logged
func (h *AuditLogsHandlers) GetActions(c echo.Context) error {
	ctx := c.Request().Context()

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get actions
	actions, err := h.auditLogsService.GetActions(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve actions")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"actions": actions,
		"count":   len(actions),
	})
}

// SoftDeleteAuditLog marks an audit log as deleted (for compliance purposes)
func (h *AuditLogsHandlers) SoftDeleteAuditLog(c echo.Context) error {
	ctx := c.Request().Context()

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get user ID from context for authorization
	userID, ok := common.GetUserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	// Parse audit log ID from URL
	auditLogID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid audit log ID")
	}

	// Check if user has permission to soft delete audit logs
	// This would typically require admin or audit manager role

	// Soft delete the audit log
	err = h.auditLogsService.SoftDeleteAuditLog(ctx, tenantID, auditLogID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to soft delete audit log")
	}

	// Log the soft delete action
	if err := h.auditLogsService.LogActivity(ctx, tenantID, "audit_logs", auditLogID.String(),
		"SOFT_DELETE", &userID, nil, nil); err != nil {
		// Log error but don't fail the request
		// This prevents infinite loops when deleting audit logs
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Audit log soft deleted successfully",
	})
}

// CreateManualAuditLog allows manual creation of audit logs (for compliance/events)
func (h *AuditLogsHandlers) CreateManualAuditLog(c echo.Context) error {
	ctx := c.Request().Context()

	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get user ID from context
	userID, ok := common.GetUserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	// Parse request body
	type ManualAuditLogRequest struct {
		TableName string       `json:"table_name" validate:"required"`
		RecordID  string       `json:"record_id" validate:"required"`
		Action    string       `json:"action" validate:"required"`
		Message   *string      `json:"message"`
		Data      models.JSONB `json:"data,omitempty"`
	}

	var req ManualAuditLogRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Create the audit log
	err := h.auditLogsService.LogActivity(ctx, tenantID, req.TableName, req.RecordID, req.Action, &userID, nil, req.Data)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create audit log")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Manual audit log created successfully",
	})
}
