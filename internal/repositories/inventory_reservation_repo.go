package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InventoryReservationRepository defines the interface for inventory reservation operations
type InventoryReservationRepository interface {
	Create(ctx context.Context, reservation *InventoryReservation) error
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*InventoryReservation, error)
	GetByReservationID(ctx context.Context, tenantID uuid.UUID, reservationID string) (*InventoryReservation, error)
	GetByProduct(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) ([]*InventoryReservation, error)
	GetByWarehouse(ctx context.Context, tenantID uuid.UUID, warehouseID uuid.UUID) ([]*InventoryReservation, error)
	GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*InventoryReservation, error)
	GetExpired(ctx context.Context, tenantID uuid.UUID) ([]*InventoryReservation, error)
	Update(ctx context.Context, reservation *InventoryReservation) error
	UpdateStatus(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, status string) error
	Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
	GetActiveReservationsByProduct(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, warehouseID uuid.UUID) ([]*InventoryReservation, error)
	GetTotalReservedQuantity(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, warehouseID uuid.UUID) (int, error)
}

// inventoryReservationRepo implements InventoryReservationRepository
type inventoryReservationRepo struct {
	db *pgxpool.Pool
}

// NewInventoryReservationRepo creates a new inventory reservation repository
func NewInventoryReservationRepo(db *pgxpool.Pool) InventoryReservationRepository {
	return &inventoryReservationRepo{db: db}
}

// Create creates a new inventory reservation
func (r *inventoryReservationRepo) Create(ctx context.Context, reservation *InventoryReservation) error {
	query := `
		INSERT INTO inventory_reservations (
			id, tenant_id, product_id, warehouse_id, reservation_id, 
			quantity, reserved_by, reserved_at, expires_at, status, 
			order_id, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	if reservation.ID == uuid.Nil {
		reservation.ID = uuid.New()
	}

	if reservation.ReservedAt.IsZero() {
		reservation.ReservedAt = time.Now()
	}

	if reservation.Status == "" {
		reservation.Status = "active"
	}

	var warehouseID *uuid.UUID
	if reservation.WarehouseID != uuid.Nil {
		warehouseID = &reservation.WarehouseID
	}

	var orderID *uuid.UUID
	if reservation.OrderID != uuid.Nil {
		orderID = &reservation.OrderID
	}

	var notes *string
	if reservation.Notes != "" {
		notes = &reservation.Notes
	}

	err := r.db.QueryRow(
		ctx, query,
		reservation.ID,
		reservation.TenantID,
		reservation.ProductID,
		warehouseID,
		reservation.ReservationID,
		reservation.Quantity,
		reservation.ReservedBy,
		reservation.ReservedAt,
		reservation.ExpiresAt,
		reservation.Status,
		orderID,
		notes,
	).Scan(&reservation.ID, &reservation.CreatedAt, &reservation.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create inventory reservation: %w", err)
	}

	return nil
}

// GetByID retrieves a reservation by ID
func (r *inventoryReservationRepo) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*InventoryReservation, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, reservation_id, 
		       quantity, reserved_by, reserved_at, expires_at, status, 
		       order_id, notes, created_at, updated_at
		FROM inventory_reservations
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	reservation := &InventoryReservation{}
	var warehouseID, orderID *uuid.UUID
	var notes *string

	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
		&reservation.ID,
		&reservation.TenantID,
		&reservation.ProductID,
		&warehouseID,
		&reservation.ReservationID,
		&reservation.Quantity,
		&reservation.ReservedBy,
		&reservation.ReservedAt,
		&reservation.ExpiresAt,
		&reservation.Status,
		&orderID,
		&notes,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	if warehouseID != nil {
		reservation.WarehouseID = *warehouseID
	}
	if orderID != nil {
		reservation.OrderID = *orderID
	}
	if notes != nil {
		reservation.Notes = *notes
	}

	return reservation, nil
}

// GetByReservationID retrieves a reservation by reservation ID
func (r *inventoryReservationRepo) GetByReservationID(ctx context.Context, tenantID uuid.UUID, reservationID string) (*InventoryReservation, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, reservation_id, 
		       quantity, reserved_by, reserved_at, expires_at, status, 
		       order_id, notes, created_at, updated_at
		FROM inventory_reservations
		WHERE reservation_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	reservation := &InventoryReservation{}
	var warehouseID, orderID *uuid.UUID
	var notes *string

	err := r.db.QueryRow(ctx, query, reservationID, tenantID).Scan(
		&reservation.ID,
		&reservation.TenantID,
		&reservation.ProductID,
		&warehouseID,
		&reservation.ReservationID,
		&reservation.Quantity,
		&reservation.ReservedBy,
		&reservation.ReservedAt,
		&reservation.ExpiresAt,
		&reservation.Status,
		&orderID,
		&notes,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get reservation by reservation ID: %w", err)
	}

	if warehouseID != nil {
		reservation.WarehouseID = *warehouseID
	}
	if orderID != nil {
		reservation.OrderID = *orderID
	}
	if notes != nil {
		reservation.Notes = *notes
	}

	return reservation, nil
}

// GetByProduct retrieves all reservations for a product
func (r *inventoryReservationRepo) GetByProduct(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) ([]*InventoryReservation, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, reservation_id, 
		       quantity, reserved_by, reserved_at, expires_at, status, 
		       order_id, notes, created_at, updated_at
		FROM inventory_reservations
		WHERE product_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		ORDER BY reserved_at DESC
	`

	rows, err := r.db.Query(ctx, query, productID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservations by product: %w", err)
	}
	defer rows.Close()

	var reservations []*InventoryReservation
	for rows.Next() {
		reservation := &InventoryReservation{}
		var warehouseID, orderID *uuid.UUID
		var notes *string

		err := rows.Scan(
			&reservation.ID,
			&reservation.TenantID,
			&reservation.ProductID,
			&warehouseID,
			&reservation.ReservationID,
			&reservation.Quantity,
			&reservation.ReservedBy,
			&reservation.ReservedAt,
			&reservation.ExpiresAt,
			&reservation.Status,
			&orderID,
			&notes,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}

		if warehouseID != nil {
			reservation.WarehouseID = *warehouseID
		}
		if orderID != nil {
			reservation.OrderID = *orderID
		}
		if notes != nil {
			reservation.Notes = *notes
		}

		reservations = append(reservations, reservation)
	}

	return reservations, nil
}

// GetByWarehouse retrieves all reservations for a warehouse
func (r *inventoryReservationRepo) GetByWarehouse(ctx context.Context, tenantID uuid.UUID, warehouseID uuid.UUID) ([]*InventoryReservation, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, reservation_id, 
		       quantity, reserved_by, reserved_at, expires_at, status, 
		       order_id, notes, created_at, updated_at
		FROM inventory_reservations
		WHERE warehouse_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		ORDER BY reserved_at DESC
	`

	rows, err := r.db.Query(ctx, query, warehouseID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservations by warehouse: %w", err)
	}
	defer rows.Close()

	var reservations []*InventoryReservation
	for rows.Next() {
		reservation := &InventoryReservation{}
		var whID, orderID *uuid.UUID
		var notes *string

		err := rows.Scan(
			&reservation.ID,
			&reservation.TenantID,
			&reservation.ProductID,
			&whID,
			&reservation.ReservationID,
			&reservation.Quantity,
			&reservation.ReservedBy,
			&reservation.ReservedAt,
			&reservation.ExpiresAt,
			&reservation.Status,
			&orderID,
			&notes,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}

		if whID != nil {
			reservation.WarehouseID = *whID
		}
		if orderID != nil {
			reservation.OrderID = *orderID
		}
		if notes != nil {
			reservation.Notes = *notes
		}

		reservations = append(reservations, reservation)
	}

	return reservations, nil
}

// GetByStatus retrieves all reservations with a specific status
func (r *inventoryReservationRepo) GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*InventoryReservation, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, reservation_id, 
		       quantity, reserved_by, reserved_at, expires_at, status, 
		       order_id, notes, created_at, updated_at
		FROM inventory_reservations
		WHERE status = $1 AND tenant_id = $2 AND deleted_at IS NULL
		ORDER BY reserved_at DESC
	`

	rows, err := r.db.Query(ctx, query, status, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservations by status: %w", err)
	}
	defer rows.Close()

	var reservations []*InventoryReservation
	for rows.Next() {
		reservation := &InventoryReservation{}
		var warehouseID, orderID *uuid.UUID
		var notes *string

		err := rows.Scan(
			&reservation.ID,
			&reservation.TenantID,
			&reservation.ProductID,
			&warehouseID,
			&reservation.ReservationID,
			&reservation.Quantity,
			&reservation.ReservedBy,
			&reservation.ReservedAt,
			&reservation.ExpiresAt,
			&reservation.Status,
			&orderID,
			&notes,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}

		if warehouseID != nil {
			reservation.WarehouseID = *warehouseID
		}
		if orderID != nil {
			reservation.OrderID = *orderID
		}
		if notes != nil {
			reservation.Notes = *notes
		}

		reservations = append(reservations, reservation)
	}

	return reservations, nil
}

// GetExpired retrieves all expired reservations
func (r *inventoryReservationRepo) GetExpired(ctx context.Context, tenantID uuid.UUID) ([]*InventoryReservation, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, reservation_id, 
		       quantity, reserved_by, reserved_at, expires_at, status, 
		       order_id, notes, created_at, updated_at
		FROM inventory_reservations
		WHERE tenant_id = $1 
		  AND status = 'active'
		  AND expires_at IS NOT NULL 
		  AND expires_at < NOW()
		  AND deleted_at IS NULL
		ORDER BY expires_at ASC
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired reservations: %w", err)
	}
	defer rows.Close()

	var reservations []*InventoryReservation
	for rows.Next() {
		reservation := &InventoryReservation{}
		var warehouseID, orderID *uuid.UUID
		var notes *string

		err := rows.Scan(
			&reservation.ID,
			&reservation.TenantID,
			&reservation.ProductID,
			&warehouseID,
			&reservation.ReservationID,
			&reservation.Quantity,
			&reservation.ReservedBy,
			&reservation.ReservedAt,
			&reservation.ExpiresAt,
			&reservation.Status,
			&orderID,
			&notes,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}

		if warehouseID != nil {
			reservation.WarehouseID = *warehouseID
		}
		if orderID != nil {
			reservation.OrderID = *orderID
		}
		if notes != nil {
			reservation.Notes = *notes
		}

		reservations = append(reservations, reservation)
	}

	return reservations, nil
}

// Update updates a reservation
func (r *inventoryReservationRepo) Update(ctx context.Context, reservation *InventoryReservation) error {
	query := `
		UPDATE inventory_reservations
		SET quantity = $1, status = $2, expires_at = $3, notes = $4, updated_at = NOW()
		WHERE id = $5 AND tenant_id = $6 AND deleted_at IS NULL
	`

	var notes *string
	if reservation.Notes != "" {
		notes = &reservation.Notes
	}

	result, err := r.db.Exec(
		ctx, query,
		reservation.Quantity,
		reservation.Status,
		reservation.ExpiresAt,
		notes,
		reservation.ID,
		reservation.TenantID,
	)

	if err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("reservation not found")
	}

	return nil
}

// UpdateStatus updates the status of a reservation
func (r *inventoryReservationRepo) UpdateStatus(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, status string) error {
	query := `
		UPDATE inventory_reservations
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND tenant_id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, status, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("reservation not found")
	}

	return nil
}

// Delete soft deletes a reservation
func (r *inventoryReservationRepo) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	query := `
		UPDATE inventory_reservations
		SET deleted_at = NOW()
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete reservation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("reservation not found")
	}

	return nil
}

// GetActiveReservationsByProduct retrieves active reservations for a product in a warehouse
func (r *inventoryReservationRepo) GetActiveReservationsByProduct(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, warehouseID uuid.UUID) ([]*InventoryReservation, error) {
	query := `
		SELECT id, tenant_id, product_id, warehouse_id, reservation_id, 
		       quantity, reserved_by, reserved_at, expires_at, status, 
		       order_id, notes, created_at, updated_at
		FROM inventory_reservations
		WHERE product_id = $1 
		  AND tenant_id = $2 
		  AND (warehouse_id = $3 OR warehouse_id IS NULL)
		  AND status = 'active'
		  AND deleted_at IS NULL
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY reserved_at DESC
	`

	rows, err := r.db.Query(ctx, query, productID, tenantID, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active reservations: %w", err)
	}
	defer rows.Close()

	var reservations []*InventoryReservation
	for rows.Next() {
		reservation := &InventoryReservation{}
		var whID, orderID *uuid.UUID
		var notes *string

		err := rows.Scan(
			&reservation.ID,
			&reservation.TenantID,
			&reservation.ProductID,
			&whID,
			&reservation.ReservationID,
			&reservation.Quantity,
			&reservation.ReservedBy,
			&reservation.ReservedAt,
			&reservation.ExpiresAt,
			&reservation.Status,
			&orderID,
			&notes,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}

		if whID != nil {
			reservation.WarehouseID = *whID
		}
		if orderID != nil {
			reservation.OrderID = *orderID
		}
		if notes != nil {
			reservation.Notes = *notes
		}

		reservations = append(reservations, reservation)
	}

	return reservations, nil
}

// GetTotalReservedQuantity calculates total reserved quantity for a product in a warehouse
func (r *inventoryReservationRepo) GetTotalReservedQuantity(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, warehouseID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(SUM(quantity), 0)
		FROM inventory_reservations
		WHERE product_id = $1 
		  AND tenant_id = $2 
		  AND (warehouse_id = $3 OR warehouse_id IS NULL)
		  AND status = 'active'
		  AND deleted_at IS NULL
		  AND (expires_at IS NULL OR expires_at > NOW())
	`

	var total int
	err := r.db.QueryRow(ctx, query, productID, tenantID, warehouseID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total reserved quantity: %w", err)
	}

	return total, nil
}
