package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry for tracking data changes
type AuditLog struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	TenantID   uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	TableName  string     `json:"table_name" db:"table_name"`
	RecordID   string     `json:"record_id" db:"record_id"`
	Action     string     `json:"action" db:"action"`
	NewValues  JSONB      `json:"new_values" db:"new_values"`
	OldValues  JSONB      `json:"old_values" db:"old_values"`
	ChangedBy  *uuid.UUID `json:"changed_by" db:"changed_by"`
	Deleted    bool       `json:"deleted" db:"deleted"`
	DeletedAt  *time.Time `json:"deleted_at" db:"deleted_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// Action constants for audit logs
const (
	ActionInsert = "INSERT"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
	ActionSoftDelete = "SOFT_DELETE"
)

// AuditLogFilters represents filters for querying audit logs
type AuditLogFilters struct {
	TableName     *string    `json:"table_name"`
	RecordID      *string    `json:"record_id"`
	Action        *string    `json:"action"`
	ChangedBy     *uuid.UUID `json:"changed_by"`
	StartDate     *time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	IncludeDeleted bool      `json:"include_deleted"`
	Limit         int        `json:"limit"`
	Offset        int        `json:"offset"`
}

// AuditLogSummary represents summary statistics for audit logs
type AuditLogSummary struct {
	TenantID         uuid.UUID         `json:"tenant_id"`
	TotalLogs        int               `json:"total_logs"`
	TableBreakdown   map[string]int    `json:"table_breakdown"`   // Table name -> count
	ActionBreakdown  map[string]int    `json:"action_breakdown"`  // Action -> count
	UserActivity     map[string]int    `json:"user_activity"`     // User ID -> count
	PeriodStart      time.Time         `json:"period_start"`
	PeriodEnd        time.Time         `json:"period_end"`
}