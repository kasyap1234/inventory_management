package repositories

import (
	"context"
	"fmt"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
)

// InventoryReservation represents inventory reservations
type InventoryReservation struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	TenantID      uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	ProductID     uuid.UUID  `json:"product_id" db:"product_id"`
	WarehouseID   uuid.UUID  `json:"warehouse_id" db:"warehouse_id"`
	ReservationID string     `json:"reservation_id" db:"reservation_id"`
	Quantity      int        `json:"quantity" db:"quantity"`
	ReservedBy    uuid.UUID  `json:"reserved_by" db:"reserved_by"`
	ReservedAt    time.Time  `json:"reserved_at" db:"reserved_at"`
	ExpiresAt     *time.Time `json:"expires_at" db:"expires_at"`
	Status        string     `json:"status" db:"status"` // active, expired, released, committed
	OrderID       uuid.UUID  `json:"order_id" db:"order_id"`
	Notes         string     `json:"notes" db:"notes"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// StockAdjustment represents stock level changes
type StockAdjustment struct {
	ID             uuid.UUID `json:"id" db:"id"`
	TenantID       uuid.UUID `json:"tenant_id" db:"tenant_id"`
	ProductID      uuid.UUID `json:"product_id" db:"product_id"`
	WarehouseID    uuid.UUID `json:"warehouse_id" db:"warehouse_id"`
	AdjustmentType string    `json:"adjustment_type" db:"adjustment_type"` // increase, decrease, reservation, release, transfer_in, transfer_out, correction, damage, return
	Quantity       int       `json:"quantity" db:"quantity"`
	PreviousStock  int       `json:"previous_stock" db:"previous_stock"`
	NewStock       int       `json:"new_stock" db:"new_stock"`
	Reason         string    `json:"reason" db:"reason"`
	ReferenceType  string    `json:"reference_type" db:"reference_type"` // order, reservation, transfer, manual
	ReferenceID    uuid.UUID `json:"reference_id" db:"reference_id"`
	AdjustedBy     uuid.UUID `json:"adjusted_by" db:"adjusted_by"`
	AdjustedAt     time.Time `json:"adjusted_at" db:"adjusted_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// InventoryAdapter wraps InventoryRepository to match services.InventoryRepository interface
type InventoryAdapter struct {
	repo                InventoryRepository
	reservationRepo     InventoryReservationRepository
	stockAdjustmentRepo StockAdjustmentRepository
}

// NewInventoryAdapter creates a new inventory adapter
func NewInventoryAdapter(repo InventoryRepository) *InventoryAdapter {
	return &InventoryAdapter{
		repo:                repo,
		reservationRepo:     nil, // Will be set via SetReservationRepo if needed
		stockAdjustmentRepo: nil, // Will be set via SetStockAdjustmentRepo if needed
	}
}

// SetReservationRepo sets the reservation repository
func (a *InventoryAdapter) SetReservationRepo(repo InventoryReservationRepository) {
	a.reservationRepo = repo
}

// SetStockAdjustmentRepo sets the stock adjustment repository
func (a *InventoryAdapter) SetStockAdjustmentRepo(repo StockAdjustmentRepository) {
	a.stockAdjustmentRepo = repo
}

// GetStock retrieves inventory by product (uses GetByProduct internally)
func (a *InventoryAdapter) GetStock(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) (*models.Inventory, error) {
	inventories, err := a.repo.GetByProduct(ctx, tenantID, productID)
	if err != nil {
		return nil, err
	}
	if len(inventories) == 0 {
		return nil, fmt.Errorf("no inventory found for product")
	}
	// Return the first inventory record (or aggregate if needed)
	return inventories[0], nil
}

// GetStockForUpdate retrieves inventory by product with SELECT FOR UPDATE lock.
// This prevents race conditions during concurrent stock reservations by acquiring
// a row-level lock that blocks other transactions from modifying the same inventory.
// MUST be called within a database transaction for the lock to be effective.
func (a *InventoryAdapter) GetStockForUpdate(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) (*models.Inventory, error) {
	inventories, err := a.repo.GetByProductForUpdate(ctx, tenantID, productID)
	if err != nil {
		return nil, err
	}
	if len(inventories) == 0 {
		return nil, fmt.Errorf("no inventory found for product")
	}
	// Return the first inventory record with lock acquired
	return inventories[0], nil
}

// UpdateStock updates the stock quantity
func (a *InventoryAdapter) UpdateStock(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, quantity int) error {
	inventories, err := a.repo.GetByProduct(ctx, tenantID, productID)
	if err != nil {
		return err
	}
	if len(inventories) == 0 {
		return fmt.Errorf("no inventory found for product")
	}

	// Update the first inventory record
	inventory := inventories[0]
	inventory.Quantity = quantity
	return a.repo.Update(ctx, inventory)
}

// UpdateReservedQuantity updates only the reserved quantity for a product
// This is separate from UpdateStock to ensure reservation updates don't accidentally modify quantity
func (a *InventoryAdapter) UpdateReservedQuantity(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, reservedQuantity int) error {
	inventories, err := a.repo.GetByProduct(ctx, tenantID, productID)
	if err != nil {
		return err
	}
	if len(inventories) == 0 {
		return fmt.Errorf("no inventory found for product")
	}

	// Update the first inventory record's reserved quantity
	inventory := inventories[0]
	inventory.ReservedQuantity = reservedQuantity
	inventory.LastUpdated = time.Now()
	return a.repo.Update(ctx, inventory)
}

// CreateReservation creates an inventory reservation
func (a *InventoryAdapter) CreateReservation(ctx context.Context, reservation *InventoryReservation) error {
	if a.reservationRepo == nil {
		return fmt.Errorf("reservation repository not configured. Please initialize the adapter with SetReservationRepo")
	}
	return a.reservationRepo.Create(ctx, reservation)
}

// GetReservation retrieves a reservation by ID
func (a *InventoryAdapter) GetReservation(ctx context.Context, tenantID uuid.UUID, reservationID string) (*InventoryReservation, error) {
	if a.reservationRepo == nil {
		return nil, fmt.Errorf("reservation repository not configured. Please initialize the adapter with SetReservationRepo")
	}
	return a.reservationRepo.GetByReservationID(ctx, tenantID, reservationID)
}

// DeleteReservation deletes a reservation
func (a *InventoryAdapter) DeleteReservation(ctx context.Context, tenantID uuid.UUID, reservationID string) error {
	if a.reservationRepo == nil {
		return fmt.Errorf("reservation repository not configured. Please initialize the adapter with SetReservationRepo")
	}
	// Get the reservation first to get its ID
	reservation, err := a.reservationRepo.GetByReservationID(ctx, tenantID, reservationID)
	if err != nil {
		return err
	}
	return a.reservationRepo.Delete(ctx, tenantID, reservation.ID)
}

// GetReservationsByProduct retrieves reservations for a product
func (a *InventoryAdapter) GetReservationsByProduct(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) ([]*InventoryReservation, error) {
	if a.reservationRepo == nil {
		return nil, fmt.Errorf("reservation repository not configured. Please initialize the adapter with SetReservationRepo")
	}
	return a.reservationRepo.GetByProduct(ctx, tenantID, productID)
}

// CreateStockAdjustment creates a stock adjustment record
func (a *InventoryAdapter) CreateStockAdjustment(ctx context.Context, adjustment *StockAdjustment) error {
	if a.stockAdjustmentRepo == nil {
		return fmt.Errorf("stock adjustment repository not configured. Please initialize the adapter with SetStockAdjustmentRepo")
	}
	return a.stockAdjustmentRepo.Create(ctx, adjustment)
}

// GetStockHistory retrieves stock adjustment history with pagination
func (a *InventoryAdapter) GetStockHistory(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, limit, offset int) ([]*StockAdjustment, error) {
	if a.stockAdjustmentRepo == nil {
		return nil, fmt.Errorf("stock adjustment repository not configured. Please initialize the adapter with SetStockAdjustmentRepo")
	}
	return a.stockAdjustmentRepo.GetByProductPaginated(ctx, tenantID, productID, limit, offset)
}

// Legacy CRUD methods - delegate to repository
func (a *InventoryAdapter) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Inventory, error) {
	return a.repo.List(ctx, tenantID, limit, offset)
}

func (a *InventoryAdapter) Create(ctx context.Context, inventory *models.Inventory) error {
	return a.repo.Create(ctx, inventory)
}

func (a *InventoryAdapter) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.Inventory, error) {
	return a.repo.GetByID(ctx, tenantID, id)
}

func (a *InventoryAdapter) Update(ctx context.Context, inventory *models.Inventory) error {
	return a.repo.Update(ctx, inventory)
}

func (a *InventoryAdapter) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	return a.repo.Delete(ctx, tenantID, id)
}

func (a *InventoryAdapter) GetByWarehouseAndProduct(ctx context.Context, tenantID uuid.UUID, warehouseID uuid.UUID, productID uuid.UUID) (*models.Inventory, error) {
	return a.repo.GetByWarehouseAndProduct(ctx, tenantID, warehouseID, productID)
}

func (a *InventoryAdapter) AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.InventorySearchFilter) ([]*models.Inventory, error) {
	return a.repo.AdvancedSearch(ctx, tenantID, filter)
}

func (a *InventoryAdapter) Transfer(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, fromWarehouseID uuid.UUID, toWarehouseID uuid.UUID, quantity int) error {
	return a.repo.Transfer(ctx, tenantID, productID, fromWarehouseID, toWarehouseID, quantity)
}

func (a *InventoryAdapter) GetByProduct(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) ([]*models.Inventory, error) {
	return a.repo.GetByProduct(ctx, tenantID, productID)
}

func (a *InventoryAdapter) GetByProductForUpdate(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) ([]*models.Inventory, error) {
	return a.repo.GetByProductForUpdate(ctx, tenantID, productID)
}

func (a *InventoryAdapter) BulkAdjust(ctx context.Context, tenantID uuid.UUID, adjustments []BulkAdjustmentItem) error {
	return a.repo.BulkAdjust(ctx, tenantID, adjustments)
}

func (a *InventoryAdapter) BulkDelete(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) error {
	return a.repo.BulkDelete(ctx, tenantID, ids)
}
