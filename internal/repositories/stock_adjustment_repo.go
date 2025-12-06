package repositories

import (
	"context"
	"fmt"
	"time"

	"agromart2/internal/common"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StockAdjustmentRepository defines the interface for stock adjustment operations
type StockAdjustmentRepository interface {
	Create(ctx context.Context, adjustment *StockAdjustment) error
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*StockAdjustment, error)
	GetByProduct(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) ([]*StockAdjustment, error)
	GetByWarehouse(ctx context.Context, tenantID uuid.UUID, warehouseID uuid.UUID) ([]*StockAdjustment, error)
	GetByProductAndWarehouse(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, warehouseID uuid.UUID) ([]*StockAdjustment, error)
	GetByAdjustmentType(ctx context.Context, tenantID uuid.UUID, adjustmentType string) ([]*StockAdjustment, error)
	GetByReference(ctx context.Context, tenantID uuid.UUID, referenceType string, referenceID uuid.UUID) ([]*StockAdjustment, error)
	GetByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*StockAdjustment, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*StockAdjustment, error)
}

// stockAdjustmentRepo implements StockAdjustmentRepository
type stockAdjustmentRepo struct {
	db *pgxpool.Pool
}

// NewStockAdjustmentRepo creates a new stock adjustment repository
func NewStockAdjustmentRepo(db *pgxpool.Pool) StockAdjustmentRepository {
	return &stockAdjustmentRepo{db: db}
}

type adjustmentRunner interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func getAdjustmentRunner(ctx context.Context, db *pgxpool.Pool) adjustmentRunner {
	if tx, ok := ctx.Value(common.TransactionKey).(pgx.Tx); ok {
		return tx
	}
	return db
}

// Create creates a new stock adjustment record
func (r *stockAdjustmentRepo) Create(ctx context.Context, adjustment *StockAdjustment) error {
	runner := getAdjustmentRunner(ctx, r.db)
	query := `
		INSERT INTO stock_adjustments (
			id, tenant_id, product_id, warehouse_id, adjustment_type,
			quantity, previous_stock, new_stock, reason, reference_type,
			reference_id, adjusted_by, adjusted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`

	if adjustment.ID == uuid.Nil {
		adjustment.ID = uuid.New()
	}

	if adjustment.AdjustedAt.IsZero() {
		adjustment.AdjustedAt = time.Now()
	}

	var warehouseID *uuid.UUID
	if adjustment.WarehouseID != uuid.Nil {
		warehouseID = &adjustment.WarehouseID
	}

	var reason *string
	if adjustment.Reason != "" {
		reason = &adjustment.Reason
	}

	var referenceType *string
	if adjustment.ReferenceType != "" {
		referenceType = &adjustment.ReferenceType
	}

	var referenceID *uuid.UUID
	if adjustment.ReferenceID != uuid.Nil {
		referenceID = &adjustment.ReferenceID
	}

	err := runner.QueryRow(
		ctx, query,
		adjustment.ID,
		adjustment.TenantID,
		adjustment.ProductID,
		warehouseID,
		adjustment.AdjustmentType,
		adjustment.Quantity,
		adjustment.PreviousStock,
		adjustment.NewStock,
		reason,
		referenceType,
		referenceID,
		adjustment.AdjustedBy,
		adjustment.AdjustedAt,
	).Scan(&adjustment.ID, &adjustment.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create stock adjustment: %w", err)
	}

	return nil
}

// GetByID retrieves a stock adjustment by ID
func (r *stockAdjustmentRepo) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*StockAdjustment, error) {
	runner := getAdjustmentRunner(ctx, r.db)
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, adjustment_type,
		       quantity, previous_stock, new_stock, reason, reference_type,
		       reference_id, adjusted_by, adjusted_at, created_at
		FROM stock_adjustments
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	adjustment := &StockAdjustment{}
	var warehouseID, referenceID *uuid.UUID
	var reason, referenceType *string

	err := runner.QueryRow(ctx, query, id, tenantID).Scan(
		&adjustment.ID,
		&adjustment.TenantID,
		&adjustment.ProductID,
		&warehouseID,
		&adjustment.AdjustmentType,
		&adjustment.Quantity,
		&adjustment.PreviousStock,
		&adjustment.NewStock,
		&reason,
		&referenceType,
		&referenceID,
		&adjustment.AdjustedBy,
		&adjustment.AdjustedAt,
		&adjustment.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get stock adjustment: %w", err)
	}

	if warehouseID != nil {
		adjustment.WarehouseID = *warehouseID
	}
	if reason != nil {
		adjustment.Reason = *reason
	}
	if referenceType != nil {
		adjustment.ReferenceType = *referenceType
	}
	if referenceID != nil {
		adjustment.ReferenceID = *referenceID
	}

	return adjustment, nil
}

// GetByProduct retrieves all stock adjustments for a product
func (r *stockAdjustmentRepo) GetByProduct(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) ([]*StockAdjustment, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, adjustment_type,
		       quantity, previous_stock, new_stock, reason, reference_type,
		       reference_id, adjusted_by, adjusted_at, created_at
		FROM stock_adjustments
		WHERE product_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		ORDER BY adjusted_at DESC
	`

	return r.scanAdjustments(ctx, query, productID, tenantID)
}

// GetByWarehouse retrieves all stock adjustments for a warehouse
func (r *stockAdjustmentRepo) GetByWarehouse(ctx context.Context, tenantID uuid.UUID, warehouseID uuid.UUID) ([]*StockAdjustment, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, adjustment_type,
		       quantity, previous_stock, new_stock, reason, reference_type,
		       reference_id, adjusted_by, adjusted_at, created_at
		FROM stock_adjustments
		WHERE warehouse_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		ORDER BY adjusted_at DESC
	`

	return r.scanAdjustments(ctx, query, warehouseID, tenantID)
}

// GetByProductAndWarehouse retrieves stock adjustments for a product in a warehouse
func (r *stockAdjustmentRepo) GetByProductAndWarehouse(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, warehouseID uuid.UUID) ([]*StockAdjustment, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, adjustment_type,
		       quantity, previous_stock, new_stock, reason, reference_type,
		       reference_id, adjusted_by, adjusted_at, created_at
		FROM stock_adjustments
		WHERE product_id = $1 AND warehouse_id = $2 AND tenant_id = $3 AND deleted_at IS NULL
		ORDER BY adjusted_at DESC
	`

	return r.scanAdjustments(ctx, query, productID, warehouseID, tenantID)
}

// GetByAdjustmentType retrieves stock adjustments by type
func (r *stockAdjustmentRepo) GetByAdjustmentType(ctx context.Context, tenantID uuid.UUID, adjustmentType string) ([]*StockAdjustment, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, adjustment_type,
		       quantity, previous_stock, new_stock, reason, reference_type,
		       reference_id, adjusted_by, adjusted_at, created_at
		FROM stock_adjustments
		WHERE adjustment_type = $1 AND tenant_id = $2 AND deleted_at IS NULL
		ORDER BY adjusted_at DESC
	`

	return r.scanAdjustments(ctx, query, adjustmentType, tenantID)
}

// GetByReference retrieves stock adjustments by reference
func (r *stockAdjustmentRepo) GetByReference(ctx context.Context, tenantID uuid.UUID, referenceType string, referenceID uuid.UUID) ([]*StockAdjustment, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, adjustment_type,
		       quantity, previous_stock, new_stock, reason, reference_type,
		       reference_id, adjusted_by, adjusted_at, created_at
		FROM stock_adjustments
		WHERE reference_type = $1 AND reference_id = $2 AND tenant_id = $3 AND deleted_at IS NULL
		ORDER BY adjusted_at DESC
	`

	return r.scanAdjustments(ctx, query, referenceType, referenceID, tenantID)
}

// GetByDateRange retrieves stock adjustments within a date range
func (r *stockAdjustmentRepo) GetByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*StockAdjustment, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, adjustment_type,
		       quantity, previous_stock, new_stock, reason, reference_type,
		       reference_id, adjusted_by, adjusted_at, created_at
		FROM stock_adjustments
		WHERE adjusted_at >= $1 AND adjusted_at <= $2 AND tenant_id = $3 AND deleted_at IS NULL
		ORDER BY adjusted_at DESC
	`

	return r.scanAdjustments(ctx, query, startDate, endDate, tenantID)
}

// List retrieves all stock adjustments with pagination
func (r *stockAdjustmentRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*StockAdjustment, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, adjustment_type,
		       quantity, previous_stock, new_stock, reason, reference_type,
		       reference_id, adjusted_by, adjusted_at, created_at
		FROM stock_adjustments
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY adjusted_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.scanAdjustments(ctx, query, tenantID, limit, offset)
}

// scanAdjustments is a helper function to scan multiple stock adjustments
func (r *stockAdjustmentRepo) scanAdjustments(ctx context.Context, query string, args ...interface{}) ([]*StockAdjustment, error) {
	runner := getAdjustmentRunner(ctx, r.db)
	rows, err := runner.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query stock adjustments: %w", err)
	}
	defer rows.Close()

	var adjustments []*StockAdjustment
	for rows.Next() {
		adjustment := &StockAdjustment{}
		var warehouseID, referenceID *uuid.UUID
		var reason, referenceType *string

		err := rows.Scan(
			&adjustment.ID,
			&adjustment.TenantID,
			&adjustment.ProductID,
			&warehouseID,
			&adjustment.AdjustmentType,
			&adjustment.Quantity,
			&adjustment.PreviousStock,
			&adjustment.NewStock,
			&reason,
			&referenceType,
			&referenceID,
			&adjustment.AdjustedBy,
			&adjustment.AdjustedAt,
			&adjustment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock adjustment: %w", err)
		}

		if warehouseID != nil {
			adjustment.WarehouseID = *warehouseID
		}
		if reason != nil {
			adjustment.Reason = *reason
		}
		if referenceType != nil {
			adjustment.ReferenceType = *referenceType
		}
		if referenceID != nil {
			adjustment.ReferenceID = *referenceID
		}

		adjustments = append(adjustments, adjustment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stock adjustments: %w", err)
	}

	return adjustments, nil
}
