package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Order, error)
	Update(ctx context.Context, order *models.Order) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Order, error)
	GetOrdersByTenantAndDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*models.Order, error)
	GetOrdersByStatus(ctx context.Context, tenantID uuid.UUID, status string, limit, offset int) ([]*models.Order, error)
	GetOrdersByTypeAndStatus(ctx context.Context, tenantID uuid.UUID, orderType, status string, limit, offset int) ([]*models.Order, error)
	GetOrdersBySupplier(ctx context.Context, tenantID uuid.UUID, supplierID uuid.UUID, limit, offset int) ([]*models.Order, error)
	GetOrdersByDistributor(ctx context.Context, tenantID uuid.UUID, distributorID uuid.UUID, limit, offset int) ([]*models.Order, error)
	AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.OrderSearchFilter) ([]*models.Order, error)
}

type orderRepo struct {
	db *pgxpool.Pool
}

func NewOrderRepo(db *pgxpool.Pool) OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO orders (id, tenant_id, order_type, supplier_id, distributor_id, product_id, warehouse_id, quantity, unit_price, status, order_date, expected_delivery, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
	`
	var supplierID, distributorID, expectedDelivery interface{}
	if order.SupplierID != nil {
		supplierID = order.SupplierID
	}
	if order.DistributorID != nil {
		distributorID = order.DistributorID
	}
	if order.ExpectedDelivery != nil {
		expectedDelivery = order.ExpectedDelivery
	} else {
		expectedDelivery = nil
	}
	_, err := r.db.Exec(ctx, query, order.ID, order.TenantID, order.OrderType, supplierID, distributorID, order.ProductID, order.WarehouseID, order.Quantity, order.UnitPrice, order.Status, order.OrderDate, expectedDelivery, order.Notes)
	return err
}

func (r *orderRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Order, error) {
	order := &models.Order{}
	query := `
		SELECT id, tenant_id, order_type, supplier_id, distributor_id, product_id, warehouse_id, quantity, unit_price, status, order_date, expected_delivery, notes, created_at, updated_at
		FROM orders
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&order.ID, &order.TenantID, &order.OrderType, &order.SupplierID, &order.DistributorID, &order.ProductID, &order.WarehouseID, &order.Quantity, &order.UnitPrice, &order.Status, &order.OrderDate, &order.ExpectedDelivery, &order.Notes, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *orderRepo) Update(ctx context.Context, order *models.Order) error {
	query := `
		UPDATE orders
		SET order_type = $1, supplier_id = $2, distributor_id = $3, product_id = $4, warehouse_id = $5, quantity = $6, unit_price = $7, status = $8, order_date = $9, expected_delivery = $10, notes = $11, updated_at = NOW()
		WHERE tenant_id = $12 AND id = $13
	`
	_, err := r.db.Exec(ctx, query, order.OrderType, order.SupplierID, order.DistributorID, order.ProductID, order.WarehouseID, order.Quantity, order.UnitPrice, order.Status, order.OrderDate, order.ExpectedDelivery, order.Notes, order.TenantID, order.ID)
	return err
}

func (r *orderRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM orders WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *orderRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	query := `
		SELECT id, tenant_id, order_type, supplier_id, distributor_id, product_id, warehouse_id, quantity, unit_price, status, order_date, expected_delivery, notes, created_at, updated_at
		FROM orders
		WHERE tenant_id = $1
		ORDER BY order_date DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		if err := rows.Scan(&order.ID, &order.TenantID, &order.OrderType, &order.SupplierID, &order.DistributorID, &order.ProductID, &order.WarehouseID, &order.Quantity, &order.UnitPrice, &order.Status, &order.OrderDate, &order.ExpectedDelivery, &order.Notes, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// AdvancedSearch performs advanced search on orders with multiple filters
func (r *orderRepo) AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.OrderSearchFilter) ([]*models.Order, error) {
	// Set defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.SortBy == "" {
		filter.SortBy = "order_date"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}

	// Build query dynamically
	queryBase := `
		SELECT o.id, o.tenant_id, o.order_type, o.supplier_id, o.distributor_id, o.product_id, o.warehouse_id, o.quantity, o.unit_price, o.status, o.order_date, o.expected_delivery, o.notes, o.created_at, o.updated_at
		FROM orders o
		WHERE o.tenant_id = $1
	`
	args := []interface{}{tenantID}
	conditionCount := 1

	// Full-text search across notes and related entities
	if filter.Query != "" {
		conditionCount++
		queryBase += fmt.Sprintf(` AND (
			COALESCE(o.notes, '') ILIKE $%d OR
			EXISTS (
				SELECT 1 FROM products p
				WHERE p.tenant_id = o.tenant_id AND p.id = o.product_id AND p.name ILIKE $%d
			) OR
			EXISTS (
				SELECT 1 FROM suppliers s
				WHERE s.tenant_id = o.tenant_id AND s.id = o.supplier_id AND s.name ILIKE $%d
			) OR
			EXISTS (
				SELECT 1 FROM distributors d
				WHERE d.tenant_id = o.tenant_id AND d.id = o.distributor_id AND d.name ILIKE $%d
			) OR
			EXISTS (
				SELECT 1 FROM warehouses w
				WHERE w.tenant_id = o.tenant_id AND w.id = o.warehouse_id AND w.name ILIKE $%d
			)
		)`, conditionCount, conditionCount, conditionCount, conditionCount, conditionCount)
		args = append(args, "%"+filter.Query+"%")
	}

	// Status filter
	if filter.Status != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND o.status = $%d`, conditionCount)
		args = append(args, *filter.Status)
	}

	// Order type filter
	if filter.OrderType != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND o.order_type = $%d`, conditionCount)
		args = append(args, *filter.OrderType)
	}

	// Supplier filter
	if filter.SupplierID != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND o.supplier_id = $%d`, conditionCount)
		args = append(args, *filter.SupplierID)
	}

	// Distributor filter
	if filter.DistributorID != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND o.distributor_id = $%d`, conditionCount)
		args = append(args, *filter.DistributorID)
	}

	// Product filter
	if filter.ProductID != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND o.product_id = $%d`, conditionCount)
		args = append(args, *filter.ProductID)
	}

	// Warehouse filter
	if filter.WarehouseID != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND o.warehouse_id = $%d`, conditionCount)
		args = append(args, *filter.WarehouseID)
	}

	// Quantity range
	if filter.MinQuantity != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND o.quantity >= $%d`, conditionCount)
		args = append(args, *filter.MinQuantity)
	}
	if filter.MaxQuantity != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND o.quantity <= $%d`, conditionCount)
		args = append(args, *filter.MaxQuantity)
	}

	// Total value range (quantity * unit_price)
	valueCondition := ""
	if filter.MinValue != nil {
		conditionCount++
		valueCondition += fmt.Sprintf(` AND (o.quantity * o.unit_price) >= $%d`, conditionCount)
		args = append(args, *filter.MinValue)
	}
	if filter.MaxValue != nil {
		conditionCount++
		valueCondition += fmt.Sprintf(` AND (o.quantity * o.unit_price) <= $%d`, conditionCount)
		args = append(args, *filter.MaxValue)
	}
	queryBase += valueCondition

	// Order date range
	if filter.OrderDateFrom != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND o.order_date >= $%d`, conditionCount)
		args = append(args, *filter.OrderDateFrom)
	}
	if filter.OrderDateTo != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND o.order_date <= $%d`, conditionCount)
		args = append(args, *filter.OrderDateTo)
	}

	// Expected delivery date range
	if filter.ExpectedDeliveryBefore != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND o.expected_delivery <= $%d`, conditionCount)
		args = append(args, *filter.ExpectedDeliveryBefore)
	}
	if filter.ExpectedDeliveryAfter != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND o.expected_delivery >= $%d`, conditionCount)
		args = append(args, *filter.ExpectedDeliveryAfter)
	}

	// Ordering
	validSortFields := map[string]bool{
		"order_date": true, "created_at": true, "quantity": true, "unit_price": true, "expected_delivery": true,
	}
	sortField := "o.order_date"
	if validSortFields[filter.SortBy] {
		sortField = "o." + filter.SortBy
	}
	sortOrder := "DESC"
	if strings.ToLower(filter.SortOrder) == "asc" {
		sortOrder = "ASC"
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

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		if err := rows.Scan(&order.ID, &order.TenantID, &order.OrderType, &order.SupplierID, &order.DistributorID, &order.ProductID, &order.WarehouseID, &order.Quantity, &order.UnitPrice, &order.Status, &order.OrderDate, &order.ExpectedDelivery, &order.Notes, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *orderRepo) GetOrdersByTenantAndDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*models.Order, error) {
	query := `
		SELECT id, tenant_id, order_type, supplier_id, distributor_id, product_id, warehouse_id, quantity, unit_price, status, order_date, expected_delivery, notes, created_at, updated_at
		FROM orders
		WHERE tenant_id = $1 AND order_date BETWEEN $2 AND $3
		ORDER BY order_date DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		if err := rows.Scan(&order.ID, &order.TenantID, &order.OrderType, &order.SupplierID, &order.DistributorID, &order.ProductID, &order.WarehouseID, &order.Quantity, &order.UnitPrice, &order.Status, &order.OrderDate, &order.ExpectedDelivery, &order.Notes, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// GetOrdersByStatus retrieves orders by status with pagination
func (r *orderRepo) GetOrdersByStatus(ctx context.Context, tenantID uuid.UUID, status string, limit, offset int) ([]*models.Order, error) {
	query := `
		SELECT id, tenant_id, order_type, supplier_id, distributor_id, product_id, warehouse_id, quantity, unit_price, status, order_date, expected_delivery, notes, created_at, updated_at
		FROM orders
		WHERE tenant_id = $1 AND status = $2
		ORDER BY order_date DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, tenantID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		if err := rows.Scan(&order.ID, &order.TenantID, &order.OrderType, &order.SupplierID, &order.DistributorID, &order.ProductID, &order.WarehouseID, &order.Quantity, &order.UnitPrice, &order.Status, &order.OrderDate, &order.ExpectedDelivery, &order.Notes, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// GetOrdersByTypeAndStatus retrieves orders by type and status with pagination
func (r *orderRepo) GetOrdersByTypeAndStatus(ctx context.Context, tenantID uuid.UUID, orderType, status string, limit, offset int) ([]*models.Order, error) {
	query := `
		SELECT id, tenant_id, order_type, supplier_id, distributor_id, product_id, warehouse_id, quantity, unit_price, status, order_date, expected_delivery, notes, created_at, updated_at
		FROM orders
		WHERE tenant_id = $1 AND order_type = $2 AND status = $3
		ORDER BY order_date DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.db.Query(ctx, query, tenantID, orderType, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		if err := rows.Scan(&order.ID, &order.TenantID, &order.OrderType, &order.SupplierID, &order.DistributorID, &order.ProductID, &order.WarehouseID, &order.Quantity, &order.UnitPrice, &order.Status, &order.OrderDate, &order.ExpectedDelivery, &order.Notes, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// GetOrdersBySupplier retrieves orders by supplier
func (r *orderRepo) GetOrdersBySupplier(ctx context.Context, tenantID uuid.UUID, supplierID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	query := `
		SELECT id, tenant_id, order_type, supplier_id, distributor_id, product_id, warehouse_id, quantity, unit_price, status, order_date, expected_delivery, notes, created_at, updated_at
		FROM orders
		WHERE tenant_id = $1 AND supplier_id = $2
		ORDER BY order_date DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, tenantID, supplierID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		if err := rows.Scan(&order.ID, &order.TenantID, &order.OrderType, &order.SupplierID, &order.DistributorID, &order.ProductID, &order.WarehouseID, &order.Quantity, &order.UnitPrice, &order.Status, &order.OrderDate, &order.ExpectedDelivery, &order.Notes, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// GetOrdersByDistributor retrieves orders by distributor
func (r *orderRepo) GetOrdersByDistributor(ctx context.Context, tenantID uuid.UUID, distributorID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	query := `
		SELECT id, tenant_id, order_type, supplier_id, distributor_id, product_id, warehouse_id, quantity, unit_price, status, order_date, expected_delivery, notes, created_at, updated_at
		FROM orders
		WHERE tenant_id = $1 AND distributor_id = $2
		ORDER BY order_date DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, tenantID, distributorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		if err := rows.Scan(&order.ID, &order.TenantID, &order.OrderType, &order.SupplierID, &order.DistributorID, &order.ProductID, &order.WarehouseID, &order.Quantity, &order.UnitPrice, &order.Status, &order.OrderDate, &order.ExpectedDelivery, &order.Notes, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}