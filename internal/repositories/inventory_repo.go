package repositories

import (
	"context"
	"fmt"
	"strings"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InventoryRepository interface {
	Create(ctx context.Context, inventory *models.Inventory) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Inventory, error)
	Update(ctx context.Context, inventory *models.Inventory) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Inventory, error)
	GetByWarehouseAndProduct(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) (*models.Inventory, error)
	AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.InventorySearchFilter) ([]*models.Inventory, error)
}

type inventoryRepo struct {
	db *pgxpool.Pool
}

func NewInventoryRepo(db *pgxpool.Pool) InventoryRepository {
	return &inventoryRepo{db: db}
}

func (r *inventoryRepo) Create(ctx context.Context, inventory *models.Inventory) error {
	query := `
		INSERT INTO inventory (id, tenant_id, warehouse_id, product_id, quantity, last_updated)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (tenant_id, warehouse_id, product_id) DO UPDATE SET quantity = inventory.quantity + EXCLUDED.quantity, last_updated = NOW()
	`
	_, err := r.db.Exec(ctx, query, inventory.ID, inventory.TenantID, inventory.WarehouseID, inventory.ProductID, inventory.Quantity)
	return err
}

func (r *inventoryRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Inventory, error) {
	inventory := &models.Inventory{}
	query := `
		SELECT id, tenant_id, warehouse_id, product_id, quantity, last_updated
		FROM inventory
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&inventory.ID, &inventory.TenantID, &inventory.WarehouseID, &inventory.ProductID, &inventory.Quantity, &inventory.LastUpdated)
	if err != nil {
		return nil, err
	}
	return inventory, nil
}

func (r *inventoryRepo) GetByWarehouseAndProduct(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) (*models.Inventory, error) {
	inventory := &models.Inventory{}
	query := `
		SELECT id, tenant_id, warehouse_id, product_id, quantity, last_updated
		FROM inventory
		WHERE tenant_id = $1 AND warehouse_id = $2 AND product_id = $3
	`
	err := r.db.QueryRow(ctx, query, tenantID, warehouseID, productID).Scan(&inventory.ID, &inventory.TenantID, &inventory.WarehouseID, &inventory.ProductID, &inventory.Quantity, &inventory.LastUpdated)
	if err != nil {
		return nil, err
	}
	return inventory, nil
}

func (r *inventoryRepo) Update(ctx context.Context, inventory *models.Inventory) error {
	query := `
		UPDATE inventory
		SET quantity = $1, last_updated = NOW()
		WHERE tenant_id = $2 AND id = $3
	`
	_, err := r.db.Exec(ctx, query, inventory.Quantity, inventory.TenantID, inventory.ID)
	return err
}

func (r *inventoryRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM inventory WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *inventoryRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Inventory, error) {
	query := `
		SELECT id, tenant_id, warehouse_id, product_id, quantity, last_updated
		FROM inventory
		WHERE tenant_id = $1
		ORDER BY last_updated DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventories []*models.Inventory
	for rows.Next() {
		inventory := &models.Inventory{}
		if err := rows.Scan(&inventory.ID, &inventory.TenantID, &inventory.WarehouseID, &inventory.ProductID, &inventory.Quantity, &inventory.LastUpdated); err != nil {
			return nil, err
		}
		inventories = append(inventories, inventory)
	}
	return inventories, nil
}

// AdvancedSearch performs advanced search on inventory with multiple filters
func (r *inventoryRepo) AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.InventorySearchFilter) ([]*models.Inventory, error) {
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
		SELECT i.id, i.tenant_id, i.warehouse_id, i.product_id, i.quantity, i.last_updated
		FROM inventory i
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
		queryBase = strings.Replace(queryBase, "FROM inventory i", "FROM inventory i LEFT JOIN products p ON p.tenant_id = i.tenant_id AND p.id = i.product_id", 1)
		sortField = "p.name"
	case "warehouse_name":
		queryBase = strings.Replace(queryBase, "FROM inventory i", "FROM inventory i LEFT JOIN warehouses w ON w.tenant_id = i.tenant_id AND w.id = i.warehouse_id", 1)
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

	rows, err := r.db.Query(ctx, queryBase, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventories []*models.Inventory
	for rows.Next() {
		inventory := &models.Inventory{}
		if err := rows.Scan(&inventory.ID, &inventory.TenantID, &inventory.WarehouseID, &inventory.ProductID, &inventory.Quantity, &inventory.LastUpdated); err != nil {
			return nil, err
		}
		inventories = append(inventories, inventory)
	}

	return inventories, nil
}