package repositories

import (
	"context"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderItemRepository interface {
	Create(ctx context.Context, orderItem *models.OrderItem) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.OrderItem, error)
	Update(ctx context.Context, orderItem *models.OrderItem) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	ListByOrderID(ctx context.Context, tenantID, orderID uuid.UUID) ([]*models.OrderItem, error)
}

type orderItemRepo struct {
	db *pgxpool.Pool
}

func NewOrderItemRepo(db *pgxpool.Pool) OrderItemRepository {
	return &orderItemRepo{db: db}
}

func (r *orderItemRepo) Create(ctx context.Context, orderItem *models.OrderItem) error {
	query := `
		INSERT INTO order_items (id, tenant_id, order_id, product_id, quantity, unit_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, orderItem.ID, orderItem.TenantID, orderItem.OrderID, orderItem.ProductID, orderItem.Quantity, orderItem.UnitPrice)
	return err
}

func (r *orderItemRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.OrderItem, error) {
	orderItem := &models.OrderItem{}
	query := `
		SELECT id, tenant_id, order_id, product_id, quantity, unit_price, created_at, updated_at
		FROM order_items
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&orderItem.ID, &orderItem.TenantID, &orderItem.OrderID, &orderItem.ProductID, &orderItem.Quantity, &orderItem.UnitPrice, &orderItem.CreatedAt, &orderItem.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return orderItem, nil
}

func (r *orderItemRepo) Update(ctx context.Context, orderItem *models.OrderItem) error {
	query := `
		UPDATE order_items
		SET quantity = $1, unit_price = $2, updated_at = NOW()
		WHERE tenant_id = $3 AND id = $4
	`
	_, err := r.db.Exec(ctx, query, orderItem.Quantity, orderItem.UnitPrice, orderItem.TenantID, orderItem.ID)
	return err
}

func (r *orderItemRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM order_items WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *orderItemRepo) ListByOrderID(ctx context.Context, tenantID, orderID uuid.UUID) ([]*models.OrderItem, error) {
	query := `
		SELECT id, tenant_id, order_id, product_id, quantity, unit_price, created_at, updated_at
		FROM order_items
		WHERE tenant_id = $1 AND order_id = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderItems []*models.OrderItem
	for rows.Next() {
		orderItem := &models.OrderItem{}
		if err := rows.Scan(&orderItem.ID, &orderItem.TenantID, &orderItem.OrderID, &orderItem.ProductID, &orderItem.Quantity, &orderItem.UnitPrice, &orderItem.CreatedAt, &orderItem.UpdatedAt); err != nil {
			return nil, err
		}
		orderItems = append(orderItems, orderItem)
	}
	return orderItems, nil
}