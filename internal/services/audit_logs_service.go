package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type AuditLogsService interface {
	// Create audit log entry
	LogActivity(ctx context.Context, tenantID uuid.UUID, tableName, recordID, action string, changedBy *uuid.UUID, oldValues, newValues models.JSONB) error

	// Query audit logs
	GetAuditLog(ctx context.Context, tenantID, auditLogID uuid.UUID) (*models.AuditLog, error)
	ListAuditLogs(ctx context.Context, tenantID uuid.UUID, filters *models.AuditLogFilters) ([]*models.AuditLog, error)

	// Get audit logs for specific entities
	GetEntityHistory(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, limit, offset int) ([]*models.AuditLog, error)
	GetUserActivity(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error)

	// Analytics and reporting
	GetAuditSummary(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*models.AuditLogSummary, error)
	GetTableNames(ctx context.Context, tenantID uuid.UUID) ([]string, error)
	GetActions(ctx context.Context, tenantID uuid.UUID) ([]string, error)

	// Timeline formatting for UI display
	FormatAuditTimeline(ctx context.Context, tenantID uuid.UUID, filters *models.AuditLogFilters) ([]TimelineGroup, error)
	FormatAuditTimelineByEntity(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, limit, offset int) ([]TimelineEvent, error)

	// Compliance operations (for audit purposes)
	SoftDeleteAuditLog(ctx context.Context, tenantID, auditLogID uuid.UUID) error

	// Helper methods for common audit scenarios
	LogEntityCreate(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, changedBy *uuid.UUID, newValues models.JSONB) error
	LogEntityUpdate(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, changedBy *uuid.UUID, oldValues, newValues models.JSONB) error
	LogEntityDelete(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, changedBy *uuid.UUID, oldValues models.JSONB) error
	LogEntitySoftDelete(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, changedBy *uuid.UUID, oldValues models.JSONB) error

	// Validation methods
	ValidateAuditFilters(filters *models.AuditLogFilters) error
}

type auditLogsService struct {
	auditLogsRepo repositories.AuditLogsRepository
}

func NewAuditLogsService(auditLogsRepo repositories.AuditLogsRepository) AuditLogsService {
	return &auditLogsService{
		auditLogsRepo: auditLogsRepo,
	}
}

// LogActivity creates a new audit log entry with validation
func (s *auditLogsService) LogActivity(ctx context.Context, tenantID uuid.UUID, tableName, recordID, action string, changedBy *uuid.UUID, oldValues, newValues models.JSONB) error {
	if tableName == "" {
		return errors.New("table_name is required")
	}
	if action == "" {
		return errors.New("action is required")
	}

	auditLog := &models.AuditLog{
		ID:        uuid.New(),
		TenantID:  tenantID,
		TableName: tableName,
		RecordID:  recordID,
		Action:    action,
		NewValues: newValues,
		OldValues: oldValues,
		ChangedBy: changedBy,
		Deleted:   false,
		DeletedAt: nil,
		CreatedAt: time.Now(),
	}

	return s.auditLogsRepo.Create(ctx, auditLog)
}

// GetAuditLog retrieves a single audit log entry
func (s *auditLogsService) GetAuditLog(ctx context.Context, tenantID, auditLogID uuid.UUID) (*models.AuditLog, error) {
	return s.auditLogsRepo.GetByID(ctx, tenantID, auditLogID)
}

// ListAuditLogs retrieves multiple audit log entries with filtering
func (s *auditLogsService) ListAuditLogs(ctx context.Context, tenantID uuid.UUID, filters *models.AuditLogFilters) ([]*models.AuditLog, error) {
	if filters == nil {
		filters = &models.AuditLogFilters{Limit: 50} // Default limit
	}
	if filters.Limit <= 0 || filters.Limit > 1000 {
		filters.Limit = 50 // Reasonable default for performance
	}

	return s.auditLogsRepo.List(ctx, tenantID, filters)
}

// GetEntityHistory retrieves audit history for a specific entity
func (s *auditLogsService) GetEntityHistory(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, limit, offset int) ([]*models.AuditLog, error) {
	return s.auditLogsRepo.GetByTableAndRecord(ctx, tenantID, tableName, recordID, limit, offset)
}

// GetUserActivity retrieves audit logs for a specific user's actions
func (s *auditLogsService) GetUserActivity(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error) {
	return s.auditLogsRepo.GetByUser(ctx, tenantID, userID, limit, offset)
}

// GetAuditSummary provides aggregated audit statistics
func (s *auditLogsService) GetAuditSummary(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*models.AuditLogSummary, error) {
	if startDate.After(endDate) {
		return nil, errors.New("start_date cannot be after end_date")
	}

	// Validate date range (not too large for performance)
	if endDate.Sub(startDate) > 365*24*time.Hour {
		return nil, errors.New("date range cannot exceed 1 year for summary queries")
	}

	return s.auditLogsRepo.GetSummary(ctx, tenantID, startDate, endDate)
}

// GetTableNames returns distinct table names that have audit logs
func (s *auditLogsService) GetTableNames(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	return s.auditLogsRepo.GetTableNames(ctx, tenantID)
}

// GetActions returns distinct actions that have been logged
func (s *auditLogsService) GetActions(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	return s.auditLogsRepo.GetActions(ctx, tenantID)
}

// SoftDeleteAuditLog marks an audit log as deleted (for compliance purposes)
func (s *auditLogsService) SoftDeleteAuditLog(ctx context.Context, tenantID, auditLogID uuid.UUID) error {
	return s.auditLogsRepo.SoftDelete(ctx, tenantID, auditLogID)
}

// Helper methods for common audit scenarios

// LogEntityCreate logs the creation of a new entity
func (s *auditLogsService) LogEntityCreate(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, changedBy *uuid.UUID, newValues models.JSONB) error {
	return s.LogActivity(ctx, tenantID, tableName, recordID, models.ActionInsert, changedBy, nil, newValues)
}

// LogEntityUpdate logs the update of an existing entity
func (s *auditLogsService) LogEntityUpdate(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, changedBy *uuid.UUID, oldValues, newValues models.JSONB) error {
	return s.LogActivity(ctx, tenantID, tableName, recordID, models.ActionUpdate, changedBy, oldValues, newValues)
}

// LogEntityDelete logs the hard deletion of an entity
func (s *auditLogsService) LogEntityDelete(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, changedBy *uuid.UUID, oldValues models.JSONB) error {
	return s.LogActivity(ctx, tenantID, tableName, recordID, models.ActionDelete, changedBy, oldValues, nil)
}

// LogEntitySoftDelete logs the soft deletion of an entity
func (s *auditLogsService) LogEntitySoftDelete(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, changedBy *uuid.UUID, oldValues models.JSONB) error {
	return s.LogActivity(ctx, tenantID, tableName, recordID, models.ActionSoftDelete, changedBy, oldValues, nil)
}

// Helper function to create JSONB for entity values (exclude sensitive fields like passwords)
func CreateEntityValues(entity interface{}) (models.JSONB, error) {
	// This is a placeholder - in a real implementation, you'd use reflection
	// or have entity-specific methods to create JSONB representations
	// You might also want to exclude sensitive fields like passwords, secrets, etc.

	switch v := entity.(type) {
	case *models.User:
		// Clean sensitive fields
		values := map[string]interface{}{
			"id":         v.ID,
			"tenant_id":  v.TenantID,
			"email":      v.Email,
			"first_name": v.FirstName,
			"last_name":  v.LastName,
			"status":     v.Status,
			"created_at": v.CreatedAt,
			"updated_at": v.UpdatedAt,
		}
		return values, nil

	case *models.Product:
		values := map[string]interface{}{
			"id":              v.ID,
			"tenant_id":       v.TenantID,
			"name":            v.Name,
			"category_id":     v.CategoryID,
			"quantity":        v.Quantity,
			"unit_price":      v.UnitPrice,
			"barcode":         v.Barcode,
			"unit_of_measure": v.UnitOfMeasure,
			"description":     v.Description,
			"batch_number":    v.BatchNumber,
			"expiry_date":     v.ExpiryDate,
			"created_at":      v.CreatedAt,
			"updated_at":      v.UpdatedAt,
		}
		return values, nil

	case *models.Order:
		values := map[string]interface{}{
			"id":                v.ID,
			"tenant_id":         v.TenantID,
			"order_type":        v.OrderType,
			"supplier_id":       v.SupplierID,
			"distributor_id":    v.DistributorID,
			"product_id":        v.ProductID,
			"warehouse_id":      v.WarehouseID,
			"quantity":          v.Quantity,
			"unit_price":        v.UnitPrice,
			"status":            v.Status,
			"order_date":        v.OrderDate,
			"expected_delivery": v.ExpectedDelivery,
			"notes":             v.Notes,
			"created_at":        v.CreatedAt,
			"updated_at":        v.UpdatedAt,
		}
		return values, nil

	default:
		return nil, fmt.Errorf("unsupported entity type for audit logging")
	}
}

// BatchLogActivities allows logging multiple activities in a single operation
func (s *auditLogsService) BatchLogActivities(ctx context.Context, tenantID uuid.UUID, logs []*models.AuditLog) error {
	// In a production system, you might want to batch insert for better performance
	// For now, we'll insert them one by one
	for _, log := range logs {
		if log.TenantID != tenantID {
			return errors.New("all audit logs must belong to the same tenant")
		}
		err := s.auditLogsRepo.Create(ctx, log)
		if err != nil {
			return fmt.Errorf("failed to create audit log for %s.%s: %w", log.TableName, log.RecordID, err)
		}
	}
	return nil
}

// TimelineEvent represents a single event in the audit timeline
type TimelineEvent struct {
	ID        uuid.UUID              `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Action    string                 `json:"action"`
	Entity    string                 `json:"entity"`
	RecordID  string                 `json:"record_id"`
	ChangedBy *uuid.UUID             `json:"changed_by"`
	Summary   string                 `json:"summary"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// TimelineGroup represents a group of events for a specific time period
type TimelineGroup struct {
	Period string          `json:"period"`
	Events []TimelineEvent `json:"events"`
}

// FormatAuditTimeline formats audit logs into a chronological timeline for UI display
func (s *auditLogsService) FormatAuditTimeline(ctx context.Context, tenantID uuid.UUID, filters *models.AuditLogFilters) ([]TimelineGroup, error) {
	if filters == nil {
		filters = &models.AuditLogFilters{Limit: 100}
	}

	// Validate filters
	if err := s.ValidateAuditFilters(filters); err != nil {
		return nil, err
	}

	// Get audit logs
	logs, err := s.ListAuditLogs(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch audit logs: %w", err)
	}

	// Convert to timeline events
	events := make([]TimelineEvent, 0, len(logs))
	for _, log := range logs {
		event := TimelineEvent{
			ID:        log.ID,
			Timestamp: log.CreatedAt,
			Action:    log.Action,
			Entity:    log.TableName,
			RecordID:  log.RecordID,
			ChangedBy: log.ChangedBy,
			Summary:   formatAuditSummary(log),
			Details:   extractChangedFields(log),
		}
		events = append(events, event)
	}

	// Sort events by timestamp (most recent first)
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[j].Timestamp.After(events[i].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	// Group events by time period
	grouped := groupEventsByPeriod(events)

	return grouped, nil
}

// FormatAuditTimelineByEntity formats audit logs for a specific entity into a timeline
func (s *auditLogsService) FormatAuditTimelineByEntity(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, limit, offset int) ([]TimelineEvent, error) {
	logs, err := s.GetEntityHistory(ctx, tenantID, tableName, recordID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch entity history: %w", err)
	}

	events := make([]TimelineEvent, 0, len(logs))
	for _, log := range logs {
		event := TimelineEvent{
			ID:        log.ID,
			Timestamp: log.CreatedAt,
			Action:    log.Action,
			Entity:    log.TableName,
			RecordID:  log.RecordID,
			ChangedBy: log.ChangedBy,
			Summary:   formatAuditSummary(log),
			Details:   extractChangedFields(log),
		}
		events = append(events, event)
	}

	// Sort by timestamp descending
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[j].Timestamp.After(events[i].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	return events, nil
}

// Helper function to format audit summary
func formatAuditSummary(log *models.AuditLog) string {
	switch log.Action {
	case models.ActionInsert:
		return fmt.Sprintf("Created %s", log.TableName)
	case models.ActionUpdate:
		return fmt.Sprintf("Updated %s", log.TableName)
	case models.ActionDelete:
		return fmt.Sprintf("Deleted %s", log.TableName)
	case models.ActionSoftDelete:
		return fmt.Sprintf("Archived %s", log.TableName)
	default:
		return fmt.Sprintf("%s %s", log.Action, log.TableName)
	}
}

// Helper function to extract changed fields from old and new values
func extractChangedFields(log *models.AuditLog) map[string]interface{} {
	details := make(map[string]interface{})

	if log.OldValues != nil {
		details["old_values"] = log.OldValues
	}
	if log.NewValues != nil {
		details["new_values"] = log.NewValues
	}

	return details
}

// Helper function to group events by time period
func groupEventsByPeriod(events []TimelineEvent) []TimelineGroup {
	if len(events) == 0 {
		return []TimelineGroup{}
	}

	groups := make(map[string][]TimelineEvent)
	periodOrder := []string{}

	for _, event := range events {
		period := formatTimePeriod(event.Timestamp)
		if _, exists := groups[period]; !exists {
			periodOrder = append(periodOrder, period)
		}
		groups[period] = append(groups[period], event)
	}

	// Create timeline groups in order
	result := make([]TimelineGroup, 0, len(groups))
	for _, period := range periodOrder {
		result = append(result, TimelineGroup{
			Period: period,
			Events: groups[period],
		})
	}

	return result
}

// Helper function to format time period
func formatTimePeriod(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < 24*time.Hour {
		return "Today"
	} else if diff < 48*time.Hour {
		return "Yesterday"
	} else if diff < 7*24*time.Hour {
		return "This Week"
	} else if diff < 30*24*time.Hour {
		return "This Month"
	} else if diff < 365*24*time.Hour {
		return t.Format("January 2006")
	}
	return t.Format("2006")
}

// ValidateAuditFilters performs security and performance validation on audit filters
func (s *auditLogsService) ValidateAuditFilters(filters *models.AuditLogFilters) error {
	if filters == nil {
		return nil
	}

	// Validate date range order
	if filters.StartDate != nil && filters.EndDate != nil {
		if filters.StartDate.After(*filters.EndDate) {
			return errors.New("start_date cannot be after end_date")
		}
		// Limit date range to prevent excessive data extraction
		if filters.EndDate.Sub(*filters.StartDate) > 365*24*time.Hour {
			return errors.New("date range cannot exceed 1 year")
		}
	}

	// Limit page size for performance
	if filters.Limit > 1000 {
		return errors.New("maximum limit is 1000 records")
	}

	return nil
}
