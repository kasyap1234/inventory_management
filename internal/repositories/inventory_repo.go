package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"agromart2/internal/common"
	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInventoryNotFound = errors.New("inventory not found")

type InventoryRepository interface {
	// Legacy CRUD methods
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Inventory, error)
	Create(ctx context.Context, inventory *models.Inventory) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Inventory, error)
	Update(ctx context.Context, inventory *models.Inventory) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	GetByWarehouseAndProduct(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) (*models.Inventory, error)
	AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.InventorySearchFilter) ([]*models.Inventory, error)
	GetByProduct(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.Inventory, error)
	// GetByProductForUpdate retrieves inventory with SELECT FOR UPDATE to prevent race conditions
	GetByProductForUpdate(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.Inventory, error)
	Transfer(ctx context.Context, tenantID, productID, fromWarehouseID, toWarehouseID uuid.UUID, quantity int) error
	// GetAnalyticsAggregates returns aggregated analytics data using SQL for performance
	GetAnalyticsAggregates(ctx context.Context, tenantID uuid.UUID) (*InventoryAnalyticsAggregates, error)
	BulkAdjust(ctx context.Context, tenantID uuid.UUID, adjustments []BulkAdjustmentItem) error
	BulkDelete(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) error
}

// BulkAdjustmentItem represents a single item in a bulk adjustment operation
type BulkAdjustmentItem struct {
	ProductID  uuid.UUID
	Quantity   int
	Reason     string
	AdjustedBy uuid.UUID
}

// InventoryAnalyticsAggregates holds pre-computed aggregates for analytics
type InventoryAnalyticsAggregates struct {
	TotalStockValue float64
	LowStockCount   int
	TotalItemsCount int
}

type inventoryRepo struct {
	db *pgxpool.Pool
}

func NewInventoryRepo(db *pgxpool.Pool) InventoryRepository {
	return &inventoryRepo{db: db}
}

// queryRunner abstracts pgxpool.Pool and pgx.Tx so repositories can transparently
// participate in caller-managed transactions (using context propagation).
type queryRunner interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func getRunner(ctx context.Context, db *pgxpool.Pool) queryRunner {
	if tx, ok := ctx.Value(common.TransactionKey).(pgx.Tx); ok {
		return tx
	}
	return db
}

func (r *inventoryRepo) Create(ctx context.Context, inventory *models.Inventory) error {
	runner := getRunner(ctx, r.db)
	query := `
		INSERT INTO inventory (id, tenant_id, warehouse_id, product_id, quantity, last_updated)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (tenant_id, warehouse_id, product_id) DO UPDATE SET quantity = inventory.quantity + EXCLUDED.quantity, last_updated = NOW()
	`
	_, err := runner.Exec(ctx, query, inventory.ID, inventory.TenantID, inventory.WarehouseID, inventory.ProductID, inventory.Quantity)
	return err
}

func (r *inventoryRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Inventory, error) {
	runner := getRunner(ctx, r.db)
	inventory := &models.Inventory{}
	query := `
		SELECT id, tenant_id, warehouse_id, product_id, quantity, last_updated
		FROM inventory
		WHERE tenant_id = $1 AND id = $2
	`
	err := runner.QueryRow(ctx, query, tenantID, id).Scan(&inventory.ID, &inventory.TenantID, &inventory.WarehouseID, &inventory.ProductID, &inventory.Quantity, &inventory.LastUpdated)
	if err != nil {
		return nil, err
	}
	return inventory, nil
}

func (r *inventoryRepo) GetByWarehouseAndProduct(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) (*models.Inventory, error) {
	runner := getRunner(ctx, r.db)
	inventory := &models.Inventory{}
	query := `
		SELECT id, tenant_id, warehouse_id, product_id, quantity, last_updated
		FROM inventory
		WHERE tenant_id = $1 AND warehouse_id = $2 AND product_id = $3
	`
	err := runner.QueryRow(ctx, query, tenantID, warehouseID, productID).Scan(&inventory.ID, &inventory.TenantID, &inventory.WarehouseID, &inventory.ProductID, &inventory.Quantity, &inventory.LastUpdated)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInventoryNotFound
		}
		return nil, err
	}
	return inventory, nil
}

func (r *inventoryRepo) Update(ctx context.Context, inventory *models.Inventory) error {
	runner := getRunner(ctx, r.db)
	query := `
		UPDATE inventory
		SET quantity = $1, reserved_quantity = $2, last_updated = NOW()
		WHERE tenant_id = $3 AND id = $4
	`
	_, err := runner.Exec(ctx, query, inventory.Quantity, inventory.ReservedQuantity, inventory.TenantID, inventory.ID)
	return err
}

func (r *inventoryRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	runner := getRunner(ctx, r.db)
	query := `DELETE FROM inventory WHERE tenant_id = $1 AND id = $2`
	_, err := runner.Exec(ctx, query, tenantID, id)
	return err
}

func (r *inventoryRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Inventory, error) {
	runner := getRunner(ctx, r.db)
	query := `
		SELECT 
			i.id, i.tenant_id, i.warehouse_id, i.product_id, i.quantity, i.last_updated,
			COALESCE(p.name, '') as product_name,
			COALESCE(w.name, '') as warehouse_name
		FROM inventory i
		LEFT JOIN products p ON p.id = i.product_id AND p.tenant_id = i.tenant_id
		LEFT JOIN warehouses w ON w.id = i.warehouse_id AND w.tenant_id = i.tenant_id
		WHERE i.tenant_id = $1
		ORDER BY i.last_updated DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := runner.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventories []*models.Inventory
	for rows.Next() {
		inventory := &models.Inventory{}
		if err := rows.Scan(
			&inventory.ID, &inventory.TenantID, &inventory.WarehouseID, &inventory.ProductID,
			&inventory.Quantity, &inventory.LastUpdated,
			&inventory.ProductName, &inventory.WarehouseName,
		); err != nil {
			return nil, err
		}
		inventories = append(inventories, inventory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating inventories: %w", err)
	}

	return inventories, nil
}

func (r *inventoryRepo) GetByProduct(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.Inventory, error) {
	runner := getRunner(ctx, r.db)
	query := `
		SELECT id, tenant_id, warehouse_id, product_id, quantity, reserved_quantity, last_updated
		FROM inventory
		WHERE tenant_id = $1 AND product_id = $2
	`

	rows, err := runner.Query(ctx, query, tenantID, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventories []*models.Inventory
	for rows.Next() {
		inventory := &models.Inventory{}
		if err := rows.Scan(&inventory.ID, &inventory.TenantID, &inventory.WarehouseID, &inventory.ProductID, &inventory.Quantity, &inventory.ReservedQuantity, &inventory.LastUpdated); err != nil {
			return nil, err
		}
		inventories = append(inventories, inventory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating inventories by product: %w", err)
	}

	return inventories, nil
}

// GetByProductForUpdate retrieves inventory with SELECT FOR UPDATE lock to prevent race conditions.
// This should be used within a transaction when performing atomic stock operations.
func (r *inventoryRepo) GetByProductForUpdate(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.Inventory, error) {
	runner := getRunner(ctx, r.db)
	query := `
		SELECT id, tenant_id, warehouse_id, product_id, quantity, reserved_quantity, last_updated
		FROM inventory
		WHERE tenant_id = $1 AND product_id = $2
		FOR UPDATE
	`

	rows, err := runner.Query(ctx, query, tenantID, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventories []*models.Inventory
	for rows.Next() {
		inventory := &models.Inventory{}
		if err := rows.Scan(&inventory.ID, &inventory.TenantID, &inventory.WarehouseID, &inventory.ProductID, &inventory.Quantity, &inventory.ReservedQuantity, &inventory.LastUpdated); err != nil {
			return nil, err
		}
		inventories = append(inventories, inventory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating inventories by product for update: %w", err)
	}

	return inventories, nil
}

// AdvancedSearch performs advanced search on inventory with multiple filters
func (r *inventoryRepo) AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.InventorySearchFilter) ([]*models.Inventory, error) {
	runner := getRunner(ctx, r.db)
	// Set defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.SortBy == "" {
		filter.SortBy = "last_updated"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}

	// Build query dynamically
	queryBase := `
		SELECT 
			i.id, i.tenant_id, i.warehouse_id, i.product_id, i.quantity, i.last_updated,
			COALESCE(p.name, '') as product_name,
			COALESCE(w.name, '') as warehouse_name
		FROM inventory i
		LEFT JOIN products p ON p.id = i.product_id AND p.tenant_id = i.tenant_id
		LEFT JOIN warehouses w ON w.id = i.warehouse_id AND w.tenant_id = i.tenant_id
		WHERE i.tenant_id = $1
	`
	args := []interface{}{tenantID}
	conditionCount := 1

	// Full-text search across product name and warehouse name
	if filter.Query != "" {
		conditionCount++
		queryBase += fmt.Sprintf(` AND (
			EXISTS (
				SELECT 1 FROM products p
				WHERE p.tenant_id = i.tenant_id AND p.id = i.product_id AND p.name ILIKE $%d
			) OR
			EXISTS (
				SELECT 1 FROM warehouses w
				WHERE w.tenant_id = i.tenant_id AND w.id = i.warehouse_id AND w.name ILIKE $%d
			)
		)`, conditionCount, conditionCount)
		args = append(args, "%"+filter.Query+"%")
	}

	// Warehouse filter
	if filter.WarehouseID != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND i.warehouse_id = $%d`, conditionCount)
		args = append(args, *filter.WarehouseID)
	}

	// Product filter
	if filter.ProductID != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND i.product_id = $%d`, conditionCount)
		args = append(args, *filter.ProductID)
	}

	// Quantity range
	if filter.MinQuantity != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND i.quantity >= $%d`, conditionCount)
		args = append(args, *filter.MinQuantity)
	}
	if filter.MaxQuantity != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND i.quantity <= $%d`, conditionCount)
		args = append(args, *filter.MaxQuantity)
	}

	// Handle MinStock and MaxStock as aliases for MinQuantity and MaxQuantity
	if filter.MinStock != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND i.quantity >= $%d`, conditionCount)
		args = append(args, *filter.MinStock)
	}
	if filter.MaxStock != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND i.quantity <= $%d`, conditionCount)
		args = append(args, *filter.MaxStock)
	}

	// Stock threshold filter (for low stock alerts)
	if filter.StockThreshold != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND i.quantity <= $%d`, conditionCount)
		args = append(args, *filter.StockThreshold)
	}

	// Last updated date range
	if filter.LastUpdatedFrom != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND i.last_updated >= $%d`, conditionCount)
		args = append(args, *filter.LastUpdatedFrom)
	}
	if filter.LastUpdatedTo != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND i.last_updated <= $%d`, conditionCount)
		args = append(args, *filter.LastUpdatedTo)
	}

	// Ordering - handle joins for sorting by product_name and warehouse_name
	sortField := "i.last_updated"
	sortOrder := "DESC"

	if strings.ToLower(filter.SortOrder) == "asc" {
		sortOrder = "ASC"
	}

	switch filter.SortBy {
	case "quantity":
		sortField = "i.quantity"
	case "last_updated":
		sortField = "i.last_updated"
	case "product_name":
		sortField = "p.name"
	case "warehouse_name":
		sortField = "w.name"
	default:
		sortField = "i.last_updated"
	}

	queryBase += fmt.Sprintf(` ORDER BY %s %s`, sortField, sortOrder)

	// Pagination
	conditionCount++
	queryBase += fmt.Sprintf(` LIMIT $%d`, conditionCount)
	args = append(args, filter.Limit)
	if filter.Offset > 0 {
		conditionCount++
		queryBase += fmt.Sprintf(` OFFSET $%d`, conditionCount)
		args = append(args, filter.Offset)
	}

	rows, err := runner.Query(ctx, queryBase, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventories []*models.Inventory
	for rows.Next() {
		inventory := &models.Inventory{}
		if err := rows.Scan(
			&inventory.ID, &inventory.TenantID, &inventory.WarehouseID, &inventory.ProductID,
			&inventory.Quantity, &inventory.LastUpdated,
			&inventory.ProductName, &inventory.WarehouseName,
		); err != nil {
			return nil, err
		}
		inventories = append(inventories, inventory)
	}

	return inventories, nil
}

// Transfer transfers stock from one warehouse to another
func (r *inventoryRepo) Transfer(ctx context.Context, tenantID, productID, fromWarehouseID, toWarehouseID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("transfer quantity must be positive")
	}

	runner := getRunner(ctx, r.db)

	// If caller already provided a transaction, reuse it; otherwise create one locally.
	tx, inTx := runner.(pgx.Tx)
	var err error
	if !inTx {
		tx, err = r.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback(ctx)
	}

	// Check if source warehouse has enough stock
	var sourceQuantity int
	err = tx.QueryRow(ctx, `
		SELECT quantity FROM inventory
		WHERE tenant_id = $1 AND warehouse_id = $2 AND product_id = $3
		FOR UPDATE
	`, tenantID, fromWarehouseID, productID).Scan(&sourceQuantity)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("source inventory not found")
		}
		return fmt.Errorf("failed to check source inventory: %w", err)
	}

	if sourceQuantity < quantity {
		return fmt.Errorf("insufficient stock in source warehouse: available %d, requested %d", sourceQuantity, quantity)
	}

	// Decrease quantity in source warehouse
	if _, err = tx.Exec(ctx, `
		UPDATE inventory
		SET quantity = quantity - $1, last_updated = NOW()
		WHERE tenant_id = $2 AND warehouse_id = $3 AND product_id = $4
	`, quantity, tenantID, fromWarehouseID, productID); err != nil {
		return fmt.Errorf("failed to decrease source inventory: %w", err)
	}

	// Increase quantity in destination warehouse (or create if doesn't exist)
	if _, err = tx.Exec(ctx, `
		INSERT INTO inventory (id, tenant_id, warehouse_id, product_id, quantity, last_updated)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (tenant_id, warehouse_id, product_id)
		DO UPDATE SET quantity = inventory.quantity + EXCLUDED.quantity, last_updated = NOW()
	`, uuid.New(), tenantID, toWarehouseID, productID, quantity); err != nil {
		return fmt.Errorf("failed to increase destination inventory: %w", err)
	}

	if !inTx {
		// Commit transaction when we created it locally
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// GetAnalyticsAggregates returns aggregated inventory analytics data using SQL
// This joins inventory with products to calculate total stock value in a single query
func (r *inventoryRepo) GetAnalyticsAggregates(ctx context.Context, tenantID uuid.UUID) (*InventoryAnalyticsAggregates, error) {
	runner := getRunner(ctx, r.db)
	query := `
		SELECT 
			COALESCE(SUM(i.quantity * p.unit_price), 0) as total_stock_value,
			COUNT(CASE WHEN i.quantity < 10 THEN 1 END) as low_stock_count,
			COUNT(*) as total_items_count
		FROM inventory i
		JOIN products p ON i.product_id = p.id AND i.tenant_id = p.tenant_id
		WHERE i.tenant_id = $1
	`

	aggregates := &InventoryAnalyticsAggregates{}
	err := runner.QueryRow(ctx, query, tenantID).Scan(
		&aggregates.TotalStockValue,
		&aggregates.LowStockCount,
		&aggregates.TotalItemsCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory analytics aggregates: %w", err)
	}

	return aggregates, nil
}

// BulkAdjust performs multiple stock adjustments in a single transaction
func (r *inventoryRepo) BulkAdjust(ctx context.Context, tenantID uuid.UUID, adjustments []BulkAdjustmentItem) error {
	if len(adjustments) == 0 {
		return nil
	}

	runner := getRunner(ctx, r.db)
	tx, inTx := runner.(pgx.Tx)
	var err error
	if !inTx {
		tx, err = r.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback(ctx)
	}

	for _, adj := range adjustments {
		// Get current stock with lock
		var currentQuantity, reservedQuantity int
		err := tx.QueryRow(ctx, `
			SELECT quantity, reserved_quantity 
			FROM inventory 
			WHERE tenant_id = $1 AND product_id = $2 
			FOR UPDATE
		`, tenantID, adj.ProductID).Scan(&currentQuantity, &reservedQuantity)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("product %s not found in inventory", adj.ProductID)
			}
			return fmt.Errorf("failed to get stock for product %s: %w", adj.ProductID, err)
		}

		newQuantity := currentQuantity + adj.Quantity
		if newQuantity < 0 {
			return fmt.Errorf("adjustment for product %s results in negative stock", adj.ProductID)
		}

		// Update inventory
		_, err = tx.Exec(ctx, `
			UPDATE inventory 
			SET quantity = $1, last_updated = NOW() 
			WHERE tenant_id = $2 AND product_id = $3
		`, newQuantity, tenantID, adj.ProductID)

		if err != nil {
			return fmt.Errorf("failed to update stock for product %s: %w", adj.ProductID, err)
		}

		// Create stock adjustment record
		adjustmentType := "increase"
		if adj.Quantity < 0 {
			adjustmentType = "decrease"
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO stock_adjustments (
				id, tenant_id, product_id, adjustment_type, quantity, 
				previous_stock, new_stock, reason, adjusted_by, adjusted_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		`, uuid.New(), tenantID, adj.ProductID, adjustmentType, adj.Quantity,
			currentQuantity, newQuantity, adj.Reason, adj.AdjustedBy)

		if err != nil {
			return fmt.Errorf("failed to create adjustment record for product %s: %w", adj.ProductID, err)
		}
	}

	if !inTx {
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// BulkDelete deletes multiple inventory records in a single transaction
func (r *inventoryRepo) BulkDelete(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	runner := getRunner(ctx, r.db)
	query := `DELETE FROM inventory WHERE tenant_id = $1 AND id = ANY($2)`
	_, err := runner.Exec(ctx, query, tenantID, ids)
	return err
}
