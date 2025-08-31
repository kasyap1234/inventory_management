package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditLogsRepository interface {
	// Create a new audit log entry
	Create(ctx context.Context, auditLog *models.AuditLog) error

	// Get audit log by ID and tenant
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.AuditLog, error)

	// List audit logs with filtering options
	List(ctx context.Context, tenantID uuid.UUID, filters *models.AuditLogFilters) ([]*models.AuditLog, error)

	// Get audit logs for a specific table and record
	GetByTableAndRecord(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, limit, offset int) ([]*models.AuditLog, error)

	// Get audit logs by user
	GetByUser(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error)

	// Soft delete an audit log (for compliance purposes)
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error

	// Get audit summary for statistics
	GetSummary(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*models.AuditLogSummary, error)

	// Get distinct table names for a tenant
	GetTableNames(ctx context.Context, tenantID uuid.UUID) ([]string, error)

	// Get distinct actions for a tenant
	GetActions(ctx context.Context, tenantID uuid.UUID) ([]string, error)
}

type auditLogsRepo struct {
	db *pgxpool.Pool
}

func NewAuditLogsRepo(db *pgxpool.Pool) AuditLogsRepository {
	return &auditLogsRepo{db: db}
}

func (r *auditLogsRepo) Create(ctx context.Context, auditLog *models.AuditLog) error {
	auditLog.CreatedAt = time.Now()
	if auditLog.ID == uuid.Nil {
		auditLog.ID = uuid.New()
	}

	query := `
		INSERT INTO audit_logs (id, tenant_id, table_name, record_id, action, new_values, old_values, changed_by, deleted, deleted_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// Marshal JSONB fields
	var newValuesBytes, oldValuesBytes []byte
	var err error

	if auditLog.NewValues != nil {
		newValuesBytes, err = json.Marshal(auditLog.NewValues)
		if err != nil {
			return fmt.Errorf("failed to marshal new_values: %w", err)
		}
	}

	if auditLog.OldValues != nil {
		oldValuesBytes, err = json.Marshal(auditLog.OldValues)
		if err != nil {
			return fmt.Errorf("failed to marshal old_values: %w", err)
		}
	}

	_, err = r.db.Exec(ctx, query,
		auditLog.ID,
		auditLog.TenantID,
		auditLog.TableName,
		auditLog.RecordID,
		auditLog.Action,
		newValuesBytes,
		oldValuesBytes,
		auditLog.ChangedBy,
		auditLog.Deleted,
		auditLog.DeletedAt,
		auditLog.CreatedAt,
	)

	return err
}

func (r *auditLogsRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.AuditLog, error) {
	auditLog := &models.AuditLog{}
	var newValuesBytes, oldValuesBytes []byte

	query := `
		SELECT id, tenant_id, table_name, record_id, action, new_values, old_values, changed_by, deleted, deleted_at, created_at
		FROM audit_logs
		WHERE tenant_id = $1 AND id = $2 AND (deleted = false OR deleted IS NULL)
	`

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&auditLog.ID,
		&auditLog.TenantID,
		&auditLog.TableName,
		&auditLog.RecordID,
		&auditLog.Action,
		&newValuesBytes,
		&oldValuesBytes,
		&auditLog.ChangedBy,
		&auditLog.Deleted,
		&auditLog.DeletedAt,
		&auditLog.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSONB fields
	if len(newValuesBytes) > 0 {
		if err := json.Unmarshal(newValuesBytes, &auditLog.NewValues); err != nil {
			return nil, fmt.Errorf("failed to unmarshal new_values: %w", err)
		}
	}

	if len(oldValuesBytes) > 0 {
		if err := json.Unmarshal(oldValuesBytes, &auditLog.OldValues); err != nil {
			return nil, fmt.Errorf("failed to unmarshal old_values: %w", err)
		}
	}

	return auditLog, nil
}

func (r *auditLogsRepo) List(ctx context.Context, tenantID uuid.UUID, filters *models.AuditLogFilters) ([]*models.AuditLog, error) {
	if filters == nil {
		filters = &models.AuditLogFilters{}
	}

	query := `
		SELECT id, tenant_id, table_name, record_id, action, new_values, old_values, changed_by, deleted, deleted_at, created_at
		FROM audit_logs
		WHERE tenant_id = $1
	`

	args := []interface{}{tenantID}
	argIdx := 1

	// Build WHERE clauses based on filters
	if filters.TableName != nil {
		argIdx++
		query += fmt.Sprintf(" AND table_name = $%d", argIdx)
		args = append(args, *filters.TableName)
	}

	if filters.RecordID != nil {
		argIdx++
		query += fmt.Sprintf(" AND record_id = $%d", argIdx)
		args = append(args, *filters.RecordID)
	}

	if filters.Action != nil {
		argIdx++
		query += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, *filters.Action)
	}

	if filters.ChangedBy != nil {
		argIdx++
		query += fmt.Sprintf(" AND changed_by = $%d", argIdx)
		args = append(args, *filters.ChangedBy)
	}

	if filters.StartDate != nil {
		argIdx++
		query += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *filters.StartDate)
	}

	if filters.EndDate != nil {
		argIdx++
		query += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, *filters.EndDate)
	}

	if !filters.IncludeDeleted {
		query += " AND (deleted = false OR deleted IS NULL)"
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		argIdx++
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filters.Limit)
		offset := filters.Offset
		if offset > 0 {
			argIdx++
			query += fmt.Sprintf(" OFFSET $%d", argIdx)
			args = append(args, offset)
		}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var auditLogs []*models.AuditLog
	for rows.Next() {
		auditLog := &models.AuditLog{}
		var newValuesBytes, oldValuesBytes []byte

		err := rows.Scan(
			&auditLog.ID,
			&auditLog.TenantID,
			&auditLog.TableName,
			&auditLog.RecordID,
			&auditLog.Action,
			&newValuesBytes,
			&oldValuesBytes,
			&auditLog.ChangedBy,
			&auditLog.Deleted,
			&auditLog.DeletedAt,
			&auditLog.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Unmarshal JSONB fields
		if len(newValuesBytes) > 0 {
			if err := json.Unmarshal(newValuesBytes, &auditLog.NewValues); err != nil {
				return nil, fmt.Errorf("failed to unmarshal new_values: %w", err)
			}
		}

		if len(oldValuesBytes) > 0 {
			if err := json.Unmarshal(oldValuesBytes, &auditLog.OldValues); err != nil {
				return nil, fmt.Errorf("failed to unmarshal old_values: %w", err)
			}
		}

		auditLogs = append(auditLogs, auditLog)
	}

	return auditLogs, nil
}

func (r *auditLogsRepo) GetByTableAndRecord(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, limit, offset int) ([]*models.AuditLog, error) {
	filters := &models.AuditLogFilters{
		TableName: &tableName,
		RecordID:  &recordID,
		Limit:     limit,
		Offset:    offset,
	}
	return r.List(ctx, tenantID, filters)
}

func (r *auditLogsRepo) GetByUser(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error) {
	filters := &models.AuditLogFilters{
		ChangedBy: &userID,
		Limit:     limit,
		Offset:    offset,
	}
	return r.List(ctx, tenantID, filters)
}

func (r *auditLogsRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `
		UPDATE audit_logs
		SET deleted = true, deleted_at = NOW()
		WHERE tenant_id = $1 AND id = $2 AND deleted = false
	`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *auditLogsRepo) GetSummary(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*models.AuditLogSummary, error) {
	// Get total count
	var totalLogs int
	query := `SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3 AND (deleted = false OR deleted IS NULL)`
	err := r.db.QueryRow(ctx, query, tenantID, startDate, endDate).Scan(&totalLogs)
	if err != nil {
		return nil, err
	}

	summary := &models.AuditLogSummary{
		TenantID:    tenantID,
		TotalLogs:   totalLogs,
		TableBreakdown: make(map[string]int),
		ActionBreakdown: make(map[string]int),
		UserActivity:    make(map[string]int),
		PeriodStart: startDate,
		PeriodEnd:   endDate,
	}

	// Get table breakdown
	tableQuery := `
		SELECT table_name, COUNT(*)
		FROM audit_logs
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3 AND (deleted = false OR deleted IS NULL)
		GROUP BY table_name
	`
	rows, err := r.db.Query(ctx, tableQuery, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		var count int
		if err := rows.Scan(&tableName, &count); err != nil {
			return nil, err
		}
		summary.TableBreakdown[tableName] = count
	}

	// Get action breakdown
	actionQuery := `
		SELECT action, COUNT(*)
		FROM audit_logs
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3 AND (deleted = false OR deleted IS NULL)
		GROUP BY action
	`
	rows, err = r.db.Query(ctx, actionQuery, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err != nil {
			return nil, err
		}
		summary.ActionBreakdown[action] = count
	}

	// Get user activity
	userQuery := `
		SELECT COALESCE(changed_by::text, 'system'), COUNT(*)
		FROM audit_logs
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3 AND (deleted = false OR deleted IS NULL)
		GROUP BY changed_by
	`
	rows, err = r.db.Query(ctx, userQuery, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		var count int
		if err := rows.Scan(&userID, &count); err != nil {
			return nil, err
		}
		summary.UserActivity[userID] = count
	}

	return summary, nil
}

func (r *auditLogsRepo) GetTableNames(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT table_name
		FROM audit_logs
		WHERE tenant_id = $1 AND (deleted = false OR deleted IS NULL)
		ORDER BY table_name
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tableNames = append(tableNames, tableName)
	}

	return tableNames, nil
}

func (r *auditLogsRepo) GetActions(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT action
		FROM audit_logs
		WHERE tenant_id = $1 AND (deleted = false OR deleted IS NULL)
		ORDER BY action
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []string
	for rows.Next() {
		var action string
		if err := rows.Scan(&action); err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}

	return actions, nil
}