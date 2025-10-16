package repositories

import (
	"context"
	"fmt"
	"agromart2/internal/models"
	"github.com/google/uuid"
	"time"
)

// InventoryReservation represents inventory reservations
type InventoryReservation struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	TenantID      uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	ProductID     uuid.UUID  `json:"product_id" db:"product_id"`
	ReservationID string     `json:"reservation_id" db:"reservation_id"`
	Quantity      int        `json:"quantity" db:"quantity"`
	ReservedBy    uuid.UUID  `json:"reserved_by" db:"reserved_by"`
	ReservedAt    time.Time  `json:"reserved_at" db:"reserved_at"`
	ExpiresAt     *time.Time `json:"expires_at" db:"expires_at"`
	Status        string     `json:"status" db:"status"` // active, expired, released
}

// StockAdjustment represents stock level changes
type StockAdjustment struct {
	ID            uuid.UUID `json:"id" db:"id"`
	TenantID      uuid.UUID `json:"tenant_id" db:"tenant_id"`
	ProductID     uuid.UUID `json:"product_id" db:"product_id"`
	AdjustmentType string   `json:"adjustment_type" db:"adjustment_type"` // increase, decrease, reservation, release
	Quantity      int       `json:"quantity" db:"quantity"`
	PreviousStock int       `json:"previous_stock" db:"previous_stock"`
	NewStock      int       `json:"new_stock" db:"new_stock"`
	Reason        string    `json:"reason" db:"reason"`
	AdjustedBy    uuid.UUID `json:"adjusted_by" db:"adjusted_by"`
	AdjustedAt    time.Time `json:"adjusted_at" db:"adjusted_at"`
}

// InventoryAdapter wraps InventoryRepository to match services.InventoryRepository interface
type InventoryAdapter struct {
	repo InventoryRepository
}

// NewInventoryAdapter creates a new inventory adapter
func NewInventoryAdapter(repo InventoryRepository) *InventoryAdapter {
	return &InventoryAdapter{repo: repo}
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

// CreateReservation creates an inventory reservation (pending feature implementation)
func (a *InventoryAdapter) CreateReservation(ctx context.Context, reservation *InventoryReservation) error {
	return fmt.Errorf("inventory reservations feature is not yet implemented. Please track quantities manually using orders and adjustments")
}

// GetReservation retrieves a reservation by ID (pending feature implementation)
func (a *InventoryAdapter) GetReservation(ctx context.Context, tenantID uuid.UUID, reservationID string) (*InventoryReservation, error) {
	return nil, fmt.Errorf("inventory reservations feature is not yet implemented. Please track quantities manually using orders and adjustments")
}

// DeleteReservation deletes a reservation (pending feature implementation)
func (a *InventoryAdapter) DeleteReservation(ctx context.Context, tenantID uuid.UUID, reservationID string) error {
	return fmt.Errorf("inventory reservations feature is not yet implemented. Please track quantities manually using orders and adjustments")
}

// GetReservationsByProduct retrieves reservations for a product (pending feature implementation)
func (a *InventoryAdapter) GetReservationsByProduct(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) ([]*InventoryReservation, error) {
	return nil, fmt.Errorf("inventory reservations feature is not yet implemented. Please track quantities manually using orders and adjustments")
}

// CreateStockAdjustment creates a stock adjustment record (pending feature implementation)
func (a *InventoryAdapter) CreateStockAdjustment(ctx context.Context, adjustment *StockAdjustment) error {
	return fmt.Errorf("stock adjustment history tracking is not yet implemented. Use the inventory adjustment API endpoints instead")
}

// GetStockHistory retrieves stock adjustment history (pending feature implementation)
func (a *InventoryAdapter) GetStockHistory(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) ([]*StockAdjustment, error) {
	return nil, fmt.Errorf("stock adjustment history tracking is not yet implemented. Use the inventory adjustment API endpoints instead")
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
