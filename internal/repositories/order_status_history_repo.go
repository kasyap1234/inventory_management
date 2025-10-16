package repositories

import (
	"context"
	"fmt"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderStatusHistoryRepository interface for order status history operations
type OrderStatusHistoryRepository interface {
	Create(ctx context.Context, history *models.OrderStatusHistory) error
	GetByOrderID(ctx context.Context, tenantID, orderID uuid.UUID) ([]*models.OrderStatusHistory, error)
	GetByUserID(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]*models.OrderStatusHistory, error)
}

// orderStatusHistoryRepo implements OrderStatusHistoryRepository
type orderStatusHistoryRepo struct {
	pool *pgxpool.Pool
}

// NewOrderStatusHistoryRepo creates a new order status history repository
func NewOrderStatusHistoryRepo(pool *pgxpool.Pool) OrderStatusHistoryRepository {
	return &orderStatusHistoryRepo{pool: pool}
}

// Create creates a new order status history record
func (r *orderStatusHistoryRepo) Create(ctx context.Context, history *models.OrderStatusHistory) error {
	query := `
		INSERT INTO order_status_history (
			id, tenant_id, order_id, old_status, new_status, 
			changed_by, notes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		history.ID,
		history.TenantID,
		history.OrderID,
		history.OldStatus,
		history.NewStatus,
		history.ChangedBy,
		history.Notes,
		history.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create order status history: %w", err)
	}

	return nil
}

// GetByOrderID retrieves order status history for a specific order
func (r *orderStatusHistoryRepo) GetByOrderID(ctx context.Context, tenantID, orderID uuid.UUID) ([]*models.OrderStatusHistory, error) {
	query := `
		SELECT 
			id, tenant_id, order_id, old_status, new_status,
			changed_by, notes, created_at
		FROM order_status_history
		WHERE tenant_id = $1 AND order_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, tenantID, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order status history: %w", err)
	}
	defer rows.Close()

	var histories []*models.OrderStatusHistory
	for rows.Next() {
		var h models.OrderStatusHistory
		err := rows.Scan(
			&h.ID,
			&h.TenantID,
			&h.OrderID,
			&h.OldStatus,
			&h.NewStatus,
			&h.ChangedBy,
			&h.Notes,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order status history: %w", err)
		}
		histories = append(histories, &h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order status history rows: %w", err)
	}

	return histories, nil
}

// GetByUserID retrieves recent order status changes made by a specific user
func (r *orderStatusHistoryRepo) GetByUserID(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]*models.OrderStatusHistory, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT 
			id, tenant_id, order_id, old_status, new_status,
			changed_by, notes, created_at
		FROM order_status_history
		WHERE tenant_id = $1 AND changed_by = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query order status history by user: %w", err)
	}
	defer rows.Close()

	var histories []*models.OrderStatusHistory
	for rows.Next() {
		var h models.OrderStatusHistory
		err := rows.Scan(
			&h.ID,
			&h.TenantID,
			&h.OrderID,
			&h.OldStatus,
			&h.NewStatus,
			&h.ChangedBy,
			&h.Notes,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order status history: %w", err)
		}
		histories = append(histories, &h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order status history rows: %w", err)
	}

	return histories, nil
}
