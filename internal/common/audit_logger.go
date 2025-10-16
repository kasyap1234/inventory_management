package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AuditAction represents the type of action being audited
type AuditAction string

const (
	AuditActionCreate AuditAction = "create"
	AuditActionRead   AuditAction = "read"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
	AuditActionLogin  AuditAction = "login"
	AuditActionLogout AuditAction = "logout"
	AuditActionExport AuditAction = "export"
	AuditActionImport AuditAction = "import"
)

// AuditEvent represents an audit event
type AuditEvent struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Action      AuditAction            `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	TableName   string                 `json:"table_name,omitempty"`
	OldValues   map[string]interface{} `json:"old_values,omitempty"`
	NewValues   map[string]interface{} `json:"new_values,omitempty"`
	Changes     map[string]interface{} `json:"changes,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Method      string                 `json:"method,omitempty"`
	Path        string                 `json:"path,omitempty"`
	StatusCode  int                    `json:"status_code,omitempty"`
	Success     bool                   `json:"success"`
	ErrorMsg    string                 `json:"error_message,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AuditLogger provides comprehensive audit logging capabilities
type AuditLogger struct {
	logger     *StructuredLogger
	repository AuditRepository
}

// AuditRepository interface for persisting audit logs
type AuditRepository interface {
	Create(ctx context.Context, event *AuditEvent) error
	List(ctx context.Context, tenantID uuid.UUID, filters AuditFilters) ([]*AuditEvent, error)
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*AuditEvent, error)
	DeleteOldEntries(ctx context.Context, olderThan time.Time) error
}

// AuditFilters represents filters for querying audit logs
type AuditFilters struct {
	UserID     *uuid.UUID   `json:"user_id,omitempty"`
	Actions    []AuditAction `json:"actions,omitempty"`
	Resources  []string     `json:"resources,omitempty"`
	StartDate  *time.Time   `json:"start_date,omitempty"`
	EndDate    *time.Time   `json:"end_date,omitempty"`
	Success    *bool        `json:"success,omitempty"`
	Limit      int          `json:"limit,omitempty"`
	Offset     int          `json:"offset,omitempty"`
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *StructuredLogger, repository AuditRepository) *AuditLogger {
	return &AuditLogger{
		logger:     logger,
		repository: repository,
	}
}

// LogCreate logs a create operation
func (a *AuditLogger) LogCreate(ctx context.Context, resource, resourceID string, newValues map[string]interface{}) {
	event := a.createBaseEvent(ctx, AuditActionCreate, resource, resourceID)
	event.NewValues = newValues
	event.Success = true
	
	a.logEvent(ctx, event)
}

// LogRead logs a read operation
func (a *AuditLogger) LogRead(ctx context.Context, resource, resourceID string) {
	event := a.createBaseEvent(ctx, AuditActionRead, resource, resourceID)
	event.Success = true
	
	a.logEvent(ctx, event)
}

// LogUpdate logs an update operation
func (a *AuditLogger) LogUpdate(ctx context.Context, resource, resourceID string, oldValues, newValues map[string]interface{}) {
	event := a.createBaseEvent(ctx, AuditActionUpdate, resource, resourceID)
	event.OldValues = oldValues
	event.NewValues = newValues
	event.Changes = a.calculateChanges(oldValues, newValues)
	event.Success = true
	
	a.logEvent(ctx, event)
}

// LogDelete logs a delete operation
func (a *AuditLogger) LogDelete(ctx context.Context, resource, resourceID string, oldValues map[string]interface{}) {
	event := a.createBaseEvent(ctx, AuditActionDelete, resource, resourceID)
	event.OldValues = oldValues
	event.Success = true
	
	a.logEvent(ctx, event)
}

// LogLogin logs a login attempt
func (a *AuditLogger) LogLogin(ctx context.Context, userID uuid.UUID, success bool, errorMsg string) {
	event := a.createBaseEvent(ctx, AuditActionLogin, "user", userID.String())
	event.UserID = &userID
	event.Success = success
	if !success {
		event.ErrorMsg = errorMsg
	}
	
	a.logEvent(ctx, event)
}

// LogLogout logs a logout
func (a *AuditLogger) LogLogout(ctx context.Context, userID uuid.UUID) {
	event := a.createBaseEvent(ctx, AuditActionLogout, "user", userID.String())
	event.UserID = &userID
	event.Success = true
	
	a.logEvent(ctx, event)
}

// LogExport logs a data export operation
func (a *AuditLogger) LogExport(ctx context.Context, resource string, filters map[string]interface{}, recordCount int) {
	event := a.createBaseEvent(ctx, AuditActionExport, resource, "")
	event.Success = true
	event.Metadata = map[string]interface{}{
		"filters":      filters,
		"record_count": recordCount,
		"export_type":  "csv", // or whatever format
	}
	
	a.logEvent(ctx, event)
}

// LogImport logs a data import operation
func (a *AuditLogger) LogImport(ctx context.Context, resource string, recordCount int, success bool, errorMsg string) {
	event := a.createBaseEvent(ctx, AuditActionImport, resource, "")
	event.Success = success
	event.Metadata = map[string]interface{}{
		"record_count": recordCount,
	}
	if !success {
		event.ErrorMsg = errorMsg
	}
	
	a.logEvent(ctx, event)
}

// LogFailedOperation logs a failed operation
func (a *AuditLogger) LogFailedOperation(ctx context.Context, action AuditAction, resource, resourceID, errorMsg string) {
	event := a.createBaseEvent(ctx, action, resource, resourceID)
	event.Success = false
	event.ErrorMsg = errorMsg
	
	a.logEvent(ctx, event)
}

// LogCustomEvent logs a custom audit event
func (a *AuditLogger) LogCustomEvent(ctx context.Context, action AuditAction, resource, resourceID string, metadata map[string]interface{}) {
	event := a.createBaseEvent(ctx, action, resource, resourceID)
	event.Success = true
	event.Metadata = metadata
	
	a.logEvent(ctx, event)
}

// createBaseEvent creates a base audit event with common fields
func (a *AuditLogger) createBaseEvent(ctx context.Context, action AuditAction, resource, resourceID string) *AuditEvent {
	event := &AuditEvent{
		ID:         uuid.New(),
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Timestamp:  time.Now(),
	}
	
	// Extract context information
	if tenantID, ok := GetTenantIDFromContext(ctx); ok {
		event.TenantID = tenantID
	}
	
	if userID, ok := GetUserIDFromContext(ctx); ok {
		event.UserID = &userID
	}
	
	// Try to get request ID from context
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			event.RequestID = id
		}
	}
	
	// Try to get HTTP context information
	if httpCtx := ctx.Value("http_context"); httpCtx != nil {
		if httpInfo, ok := httpCtx.(map[string]interface{}); ok {
			if ipAddress, ok := httpInfo["ip_address"].(string); ok {
				event.IPAddress = ipAddress
			}
			if userAgent, ok := httpInfo["user_agent"].(string); ok {
				event.UserAgent = userAgent
			}
			if method, ok := httpInfo["method"].(string); ok {
				event.Method = method
			}
			if path, ok := httpInfo["path"].(string); ok {
				event.Path = path
			}
			if statusCode, ok := httpInfo["status_code"].(int); ok {
				event.StatusCode = statusCode
			}
		}
	}
	
	return event
}

// logEvent logs the audit event both to structured logs and persistent storage
func (a *AuditLogger) logEvent(ctx context.Context, event *AuditEvent) {
	// Log to structured logger
	a.logger.LogAudit(ctx, string(event.Action), event.Resource, event.ResourceID, event.Changes)
	
	// Persist to database
	if a.repository != nil {
		if err := a.repository.Create(ctx, event); err != nil {
			a.logger.ErrorWithContext(ctx, "Failed to persist audit event", err, map[string]interface{}{
				"audit_event_id": event.ID,
				"action":         event.Action,
				"resource":       event.Resource,
			})
		}
	}
}

// calculateChanges calculates the differences between old and new values
func (a *AuditLogger) calculateChanges(oldValues, newValues map[string]interface{}) map[string]interface{} {
	changes := make(map[string]interface{})
	
	// Check for new or changed fields
	for key, newValue := range newValues {
		if oldValue, exists := oldValues[key]; exists {
			if !a.valuesEqual(oldValue, newValue) {
				changes[key] = map[string]interface{}{
					"old": oldValue,
					"new": newValue,
				}
			}
		} else {
			changes[key] = map[string]interface{}{
				"old": nil,
				"new": newValue,
			}
		}
	}
	
	// Check for removed fields
	for key, oldValue := range oldValues {
		if _, exists := newValues[key]; !exists {
			changes[key] = map[string]interface{}{
				"old": oldValue,
				"new": nil,
			}
		}
	}
	
	return changes
}

// valuesEqual compares two values for equality
func (a *AuditLogger) valuesEqual(v1, v2 interface{}) bool {
	// Convert both values to JSON for comparison
	// This handles complex types and ensures consistent comparison
	json1, err1 := json.Marshal(v1)
	json2, err2 := json.Marshal(v2)
	
	if err1 != nil || err2 != nil {
		// Fallback to string comparison if JSON marshaling fails
		return fmt.Sprintf("%v", v1) == fmt.Sprintf("%v", v2)
	}
	
	return string(json1) == string(json2)
}

// GetAuditLogs retrieves audit logs with filters
func (a *AuditLogger) GetAuditLogs(ctx context.Context, tenantID uuid.UUID, filters AuditFilters) ([]*AuditEvent, error) {
	if a.repository == nil {
		return nil, fmt.Errorf("audit repository not configured")
	}
	
	return a.repository.List(ctx, tenantID, filters)
}

// GetAuditLog retrieves a specific audit log entry
func (a *AuditLogger) GetAuditLog(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*AuditEvent, error) {
	if a.repository == nil {
		return nil, fmt.Errorf("audit repository not configured")
	}
	
	return a.repository.GetByID(ctx, tenantID, id)
}

// CleanupOldLogs removes audit logs older than the specified duration
func (a *AuditLogger) CleanupOldLogs(ctx context.Context, retentionPeriod time.Duration) error {
	if a.repository == nil {
		return fmt.Errorf("audit repository not configured")
	}
	
	cutoffDate := time.Now().Add(-retentionPeriod)
	return a.repository.DeleteOldEntries(ctx, cutoffDate)
}

// Global audit logger instance
var globalAuditLogger *AuditLogger

// SetGlobalAuditLogger sets the global audit logger instance
func SetGlobalAuditLogger(logger *AuditLogger) {
	globalAuditLogger = logger
}

// GetGlobalAuditLogger returns the global audit logger instance
func GetGlobalAuditLogger() *AuditLogger {
	return globalAuditLogger
}

// Convenience functions for global audit logging
func AuditCreate(ctx context.Context, resource, resourceID string, newValues map[string]interface{}) {
	if globalAuditLogger != nil {
		globalAuditLogger.LogCreate(ctx, resource, resourceID, newValues)
	}
}

func AuditUpdate(ctx context.Context, resource, resourceID string, oldValues, newValues map[string]interface{}) {
	if globalAuditLogger != nil {
		globalAuditLogger.LogUpdate(ctx, resource, resourceID, oldValues, newValues)
	}
}

func AuditDelete(ctx context.Context, resource, resourceID string, oldValues map[string]interface{}) {
	if globalAuditLogger != nil {
		globalAuditLogger.LogDelete(ctx, resource, resourceID, oldValues)
	}
}