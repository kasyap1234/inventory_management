package services

import (
	"context"
	"fmt"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// InventoryRepository interface for inventory operations
type InventoryRepository interface {
	GetStock(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) (*models.Inventory, error)
	// GetStockForUpdate retrieves stock with SELECT FOR UPDATE lock for atomic reservation operations
	// This prevents race conditions during concurrent stock reservations
	GetStockForUpdate(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) (*models.Inventory, error)
	UpdateStock(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, quantity int) error
	UpdateReservedQuantity(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, reservedQuantity int) error
	CreateReservation(ctx context.Context, reservation *repositories.InventoryReservation) error
	GetReservation(ctx context.Context, tenantID uuid.UUID, reservationID string) (*repositories.InventoryReservation, error)
	DeleteReservation(ctx context.Context, tenantID uuid.UUID, reservationID string) error
	GetReservationsByProduct(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) ([]*repositories.InventoryReservation, error)
	CreateStockAdjustment(ctx context.Context, adjustment *repositories.StockAdjustment) error
	GetStockHistory(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) ([]*repositories.StockAdjustment, error)
	GetStockHistoryPaginated(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, limit, offset int) ([]*repositories.StockAdjustment, error)
	// Legacy CRUD methods
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Inventory, error)
	Create(ctx context.Context, inventory *models.Inventory) error
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.Inventory, error)
	Update(ctx context.Context, inventory *models.Inventory) error
	Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
	GetByWarehouseAndProduct(ctx context.Context, tenantID uuid.UUID, warehouseID uuid.UUID, productID uuid.UUID) (*models.Inventory, error)
	AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.InventorySearchFilter) ([]*models.Inventory, error)
	Transfer(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, fromWarehouseID uuid.UUID, toWarehouseID uuid.UUID, quantity int) error
	BulkAdjust(ctx context.Context, tenantID uuid.UUID, adjustments []repositories.BulkAdjustmentItem) error
	BulkDelete(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) error
}

// InventoryRecord represents current inventory levels (alias for models.Inventory)
type InventoryRecord = models.Inventory

// Type aliases for convenience
type InventoryReservation = repositories.InventoryReservation
type StockAdjustment = repositories.StockAdjustment

// InventoryService interface for inventory business logic operations
type InventoryService interface {
	ReserveStock(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, quantity int, reservationID string) error
	ReleaseStock(ctx context.Context, tenantID uuid.UUID, reservationID string) error
	CommitStock(ctx context.Context, tenantID uuid.UUID, reservationID string) error
	AdjustStock(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, quantity int, reason string, adjustedBy uuid.UUID) error
	GetAvailableStock(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) (int, error)
	GetByWarehouseAndProduct(ctx context.Context, tenantID uuid.UUID, warehouseID uuid.UUID, productID uuid.UUID) (*InventoryRecord, error)
	Transfer(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, fromWarehouseID uuid.UUID, toWarehouseID uuid.UUID, quantity int) error
	AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.InventorySearchFilter) ([]*InventoryRecord, error)
	GetInventoryHistory(ctx context.Context, tenantID uuid.UUID, inventoryID uuid.UUID, limit, offset int) ([]*models.AuditLog, error)
	BulkAdjustStock(ctx context.Context, tenantID uuid.UUID, adjustments []repositories.BulkAdjustmentItem) error
	BulkDelete(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) error
	// Legacy methods for handlers
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*InventoryRecord, error)
	Create(ctx context.Context, inventory *InventoryRecord) error
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*InventoryRecord, error)
	Update(ctx context.Context, inventory *InventoryRecord) error
	Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
}

// inventoryService implements InventoryService
type inventoryService struct {
	repository         InventoryRepository
	logger             *common.StructuredLogger
	transactionManager *common.TransactionManager
}

// NewInventoryService creates a new inventory service instance
func NewInventoryService(repository InventoryRepository, logger *common.StructuredLogger, txManager *common.TransactionManager) InventoryService {
	return &inventoryService{
		repository:         repository,
		logger:             logger,
		transactionManager: txManager,
	}
}

// runWithTx executes the given function within a transaction when a transaction manager is configured.
// This ensures SELECT ... FOR UPDATE locks are held for the full reservation/release lifecycle.
func (s *inventoryService) runWithTx(ctx context.Context, fn func(context.Context) error) error {
	if s.transactionManager == nil {
		return fn(ctx)
	}

	return s.transactionManager.ExecuteInTransaction(ctx, func(txCtx context.Context, tx pgx.Tx) error {
		return fn(txCtx)
	})
}

// GetByWarehouseAndProduct retrieves inventory by warehouse and product
func (s *inventoryService) GetByWarehouseAndProduct(ctx context.Context, tenantID uuid.UUID, warehouseID uuid.UUID, productID uuid.UUID) (*InventoryRecord, error) {
	return s.repository.GetByWarehouseAndProduct(ctx, tenantID, warehouseID, productID)
}

// Transfer transfers stock between warehouses
func (s *inventoryService) Transfer(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, fromWarehouseID uuid.UUID, toWarehouseID uuid.UUID, quantity int) error {
	return s.repository.Transfer(ctx, tenantID, productID, fromWarehouseID, toWarehouseID, quantity)
}

// AdvancedSearch performs advanced search on inventory records
func (s *inventoryService) AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.InventorySearchFilter) ([]*InventoryRecord, error) {
	return s.repository.AdvancedSearch(ctx, tenantID, filter)
}

// GetInventoryHistory retrieves audit history for inventory records
// Uses the stock adjustments repository to get actual change history
func (s *inventoryService) GetInventoryHistory(ctx context.Context, tenantID uuid.UUID, inventoryID uuid.UUID, limit, offset int) ([]*models.AuditLog, error) {
	// Get the inventory record to find the product ID
	inventory, err := s.repository.GetByID(ctx, tenantID, inventoryID)
	if err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to get inventory for history lookup", err, map[string]interface{}{
			"inventory_id": inventoryID,
		})
		return nil, common.CreateDatabaseError("get_inventory_history", err)
	}

	// Get paginated stock adjustment history for this product using database-level pagination
	// This is more efficient than fetching all records and slicing in memory
	adjustments, err := s.repository.GetStockHistoryPaginated(ctx, tenantID, inventory.ProductID, limit, offset)
	if err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to get stock history", err, map[string]interface{}{
			"product_id": inventory.ProductID,
		})
		return nil, common.CreateDatabaseError("get_inventory_history", err)
	}

	// Convert stock adjustments to audit log format for consistent API response
	auditLogs := make([]*models.AuditLog, 0, len(adjustments))
	for _, adj := range adjustments {
		adjustedBy := adj.AdjustedBy // Copy to take address
		auditLog := &models.AuditLog{
			ID:        adj.ID,
			TenantID:  adj.TenantID,
			TableName: "inventory",
			RecordID:  inventoryID.String(),
			Action:    adj.AdjustmentType,
			NewValues: map[string]interface{}{
				"quantity":       adj.NewStock,
				"adjustment":     adj.Quantity,
				"reason":         adj.Reason,
				"reference_type": adj.ReferenceType,
			},
			OldValues: map[string]interface{}{
				"quantity": adj.PreviousStock,
			},
			ChangedBy: &adjustedBy,
			CreatedAt: adj.AdjustedAt,
		}
		auditLogs = append(auditLogs, auditLog)
	}

	return auditLogs, nil
}

// ReserveStock reserves stock for a specific reservation using transactional semantics.
// Uses SELECT FOR UPDATE to prevent race conditions during concurrent reservations.
func (s *inventoryService) ReserveStock(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, quantity int, reservationID string) error {
	return s.runWithTx(ctx, func(txCtx context.Context) error {
		return s.reserveStockInternal(txCtx, tenantID, productID, quantity, reservationID)
	})
}

func (s *inventoryService) reserveStockInternal(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, quantity int, reservationID string) error {
	// Validate inputs
	if quantity <= 0 {
		return common.CreateValidationError("reserve_stock", map[string]interface{}{
			"quantity": "Quantity must be positive",
		})
	}
	if reservationID == "" {
		return common.CreateValidationError("reserve_stock", map[string]interface{}{
			"reservation_id": "Reservation ID is required",
		})
	}

	// Get current stock with FOR UPDATE lock to prevent concurrent modifications
	// This ensures atomicity between the stock check and reservation creation
	stock, err := s.repository.GetStockForUpdate(ctx, tenantID, productID)
	if err != nil {
		return common.CreateDatabaseError("reserve_stock", err)
	}

	// Validate stock exists and is not nil
	if stock == nil {
		return common.CreateValidationError("reserve_stock", map[string]interface{}{
			"product": "Product not found in inventory",
		})
	}

	// Check if enough stock is available
	availableStock := stock.Quantity - stock.ReservedQuantity
	if availableStock < 0 {
		availableStock = 0
	}
	if availableStock < quantity {
		return common.CreateValidationError("reserve_stock", map[string]interface{}{
			"stock": fmt.Sprintf("Insufficient stock. Available: %d, Requested: %d", availableStock, quantity),
		})
	}

	// Get user ID from context
	userID, _ := common.GetUserIDFromContext(ctx)

	// Create reservation
	reservation := &InventoryReservation{
		ID:            uuid.New(),
		TenantID:      tenantID,
		ProductID:     productID,
		ReservationID: reservationID,
		Quantity:      quantity,
		ReservedBy:    userID,
		ReservedAt:    time.Now(),
		ExpiresAt:     nil, // No expiration for order reservations
		Status:        "active",
	}

	if err := s.repository.CreateReservation(ctx, reservation); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to create inventory reservation", err, map[string]interface{}{
			"product_id":     productID,
			"quantity":       quantity,
			"reservation_id": reservationID,
		})
		return common.CreateDatabaseError("reserve_stock", err)
	}

	// Update reserved quantity
	newReservedQuantity := stock.ReservedQuantity + quantity
	if err := s.updateReservedQuantity(ctx, tenantID, productID, newReservedQuantity); err != nil {
		return err
	}

	// Create stock adjustment record
	adjustment := &StockAdjustment{
		ID:             uuid.New(),
		TenantID:       tenantID,
		ProductID:      productID,
		AdjustmentType: "reservation",
		Quantity:       quantity,
		PreviousStock:  stock.ReservedQuantity,
		NewStock:       newReservedQuantity,
		Reason:         fmt.Sprintf("Reserved for %s", reservationID),
		AdjustedBy:     userID,
		AdjustedAt:     time.Now(),
	}

	if err := s.repository.CreateStockAdjustment(ctx, adjustment); err != nil {
		s.logger.WarnWithContext(ctx, "Failed to create stock adjustment record", map[string]interface{}{
			"product_id": productID,
			"error":      err.Error(),
		})
	}

	s.logger.InfoWithContext(ctx, "Stock reserved successfully", map[string]interface{}{
		"product_id":      productID,
		"quantity":        quantity,
		"reservation_id":  reservationID,
		"available_after": availableStock - quantity,
	})

	// Check for low stock alerts
	s.checkLowStockAlert(ctx, tenantID, productID, availableStock-quantity)

	return nil
}

// ReleaseStock releases a stock reservation (alias for backward compatibility)
func (s *inventoryService) ReleaseStock(ctx context.Context, tenantID uuid.UUID, reservationID string) error {
	return s.ReleaseReservation(ctx, tenantID, reservationID)
}

// ReleaseReservation releases a stock reservation with transactional safety.
func (s *inventoryService) ReleaseReservation(ctx context.Context, tenantID uuid.UUID, reservationID string) error {
	return s.runWithTx(ctx, func(txCtx context.Context) error {
		// Get reservation
		reservation, err := s.repository.GetReservation(txCtx, tenantID, reservationID)
		if err != nil {
			return common.CreateDatabaseError("release_reservation", err)
		}

		if reservation.Status != "active" {
			return common.CreateValidationError("release_reservation", map[string]interface{}{
				"reservation": "Reservation is not active",
			})
		}

		// Get current stock
		stock, err := s.repository.GetStock(txCtx, tenantID, reservation.ProductID)
		if err != nil {
			return common.CreateDatabaseError("release_reservation", err)
		}

		// Update reserved quantity
		newReservedQuantity := stock.ReservedQuantity - reservation.Quantity
		if newReservedQuantity < 0 {
			newReservedQuantity = 0
		}

		if err := s.updateReservedQuantity(txCtx, tenantID, reservation.ProductID, newReservedQuantity); err != nil {
			return err
		}

		// Delete reservation
		if err := s.repository.DeleteReservation(txCtx, tenantID, reservationID); err != nil {
			s.logger.ErrorWithContext(txCtx, "Failed to delete reservation", err, map[string]interface{}{
				"reservation_id": reservationID,
			})
			return common.CreateDatabaseError("release_reservation", err)
		}

		// Get user ID from context
		userID, _ := common.GetUserIDFromContext(txCtx)

		// Create stock adjustment record
		adjustment := &StockAdjustment{
			ID:             uuid.New(),
			TenantID:       tenantID,
			ProductID:      reservation.ProductID,
			AdjustmentType: "release",
			Quantity:       reservation.Quantity,
			PreviousStock:  stock.ReservedQuantity,
			NewStock:       newReservedQuantity,
			Reason:         fmt.Sprintf("Released reservation %s", reservationID),
			AdjustedBy:     userID,
			AdjustedAt:     time.Now(),
		}

		if err := s.repository.CreateStockAdjustment(txCtx, adjustment); err != nil {
			s.logger.WarnWithContext(txCtx, "Failed to create stock adjustment record", map[string]interface{}{
				"product_id": reservation.ProductID,
				"error":      err.Error(),
			})
		}

		s.logger.InfoWithContext(txCtx, "Reservation released successfully", map[string]interface{}{
			"product_id":     reservation.ProductID,
			"quantity":       reservation.Quantity,
			"reservation_id": reservationID,
		})

		return nil
	})
}

// CommitStock commits a reservation (deducts from available stock)
func (s *inventoryService) CommitStock(ctx context.Context, tenantID uuid.UUID, reservationID string) error {
	return s.runWithTx(ctx, func(txCtx context.Context) error {
		// Get reservation
		reservation, err := s.repository.GetReservation(txCtx, tenantID, reservationID)
		if err != nil {
			return common.CreateDatabaseError("commit_reservation", err)
		}

		if reservation.Status != "active" {
			return common.CreateValidationError("commit_reservation", map[string]interface{}{
				"reservation": "Reservation is not active",
			})
		}

		// Delete reservation (committed reservations are removed)
		if err := s.repository.DeleteReservation(txCtx, tenantID, reservationID); err != nil {
			return common.CreateDatabaseError("commit_reservation", err)
		}

		return nil
	})
}

// AdjustStock adjusts stock levels (increase or decrease)
func (s *inventoryService) AdjustStock(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, adjustment int, reason string, adjustedBy uuid.UUID) error {
	// Validate inputs
	if adjustment == 0 {
		return common.CreateValidationError("adjust_stock", map[string]interface{}{
			"adjustment": "Adjustment cannot be zero",
		})
	}
	if reason == "" {
		reason = "Manual adjustment"
	}

	// Get current stock
	stock, err := s.repository.GetStock(ctx, tenantID, productID)
	if err != nil {
		return common.CreateDatabaseError("adjust_stock", err)
	}

	// Validate stock exists
	if stock == nil {
		return common.CreateValidationError("adjust_stock", map[string]interface{}{
			"product": "Product not found in inventory",
		})
	}

	// Calculate new stock level
	newQuantity := stock.Quantity + adjustment
	if newQuantity < 0 {
		return common.CreateValidationError("adjust_stock", map[string]interface{}{
			"stock": fmt.Sprintf("Adjustment would result in negative stock. Current: %d, Adjustment: %d", stock.Quantity, adjustment),
		})
	}

	// Check if adjustment would make available stock negative
	availableStock := newQuantity - stock.ReservedQuantity
	if availableStock < 0 {
		return common.CreateValidationError("adjust_stock", map[string]interface{}{
			"stock": fmt.Sprintf("Adjustment would result in negative available stock. Reserved: %d, New Total: %d", stock.ReservedQuantity, newQuantity),
		})
	}

	// Update stock
	if err := s.repository.UpdateStock(ctx, tenantID, productID, newQuantity); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to update stock", err, map[string]interface{}{
			"product_id":   productID,
			"adjustment":   adjustment,
			"new_quantity": newQuantity,
		})
		return common.CreateDatabaseError("adjust_stock", err)
	}

	// Create stock adjustment record
	adjustmentType := "increase"
	if adjustment < 0 {
		adjustmentType = "decrease"
	}

	stockAdjustment := &StockAdjustment{
		ID:             uuid.New(),
		TenantID:       tenantID,
		ProductID:      productID,
		AdjustmentType: adjustmentType,
		Quantity:       adjustment,
		PreviousStock:  stock.Quantity,
		NewStock:       newQuantity,
		Reason:         reason,
		AdjustedBy:     adjustedBy,
		AdjustedAt:     time.Now(),
	}

	if err := s.repository.CreateStockAdjustment(ctx, stockAdjustment); err != nil {
		s.logger.WarnWithContext(ctx, "Failed to create stock adjustment record", map[string]interface{}{
			"product_id": productID,
			"error":      err.Error(),
		})
	}

	s.logger.InfoWithContext(ctx, "Stock adjusted successfully", map[string]interface{}{
		"product_id":     productID,
		"adjustment":     adjustment,
		"previous_stock": stock.Quantity,
		"new_stock":      newQuantity,
		"reason":         reason,
	})

	// Check for low stock alerts
	availableStock = newQuantity - stock.ReservedQuantity
	s.checkLowStockAlert(ctx, tenantID, productID, availableStock)

	// Audit log
	common.AuditUpdate(ctx, "inventory", productID.String(),
		map[string]interface{}{"quantity": stock.Quantity},
		map[string]interface{}{"quantity": newQuantity, "reason": reason},
	)

	return nil
}

// GetAvailableStock returns available (non-reserved) stock
func (s *inventoryService) GetAvailableStock(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) (int, error) {
	stock, err := s.repository.GetStock(ctx, tenantID, productID)
	if err != nil {
		return 0, common.CreateDatabaseError("get_available_stock", err)
	}

	// Handle nil stock
	if stock == nil {
		return 0, common.CreateValidationError("get_available_stock", map[string]interface{}{
			"product": "Product not found in inventory",
		})
	}

	availableStock := stock.Quantity - stock.ReservedQuantity
	if availableStock < 0 {
		// Log inconsistency but return 0
		s.logger.WarnWithContext(ctx, "Negative available stock detected", map[string]interface{}{
			"product_id":        productID,
			"quantity":          stock.Quantity,
			"reserved_quantity": stock.ReservedQuantity,
		})
		availableStock = 0
	}

	return availableStock, nil
}

// GetReservedStock returns reserved stock quantity
func (s *inventoryService) GetReservedStock(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID) (int, error) {
	stock, err := s.repository.GetStock(ctx, tenantID, productID)
	if err != nil {
		return 0, common.CreateDatabaseError("get_reserved_stock", err)
	}

	return stock.ReservedQuantity, nil
}

// Helper methods

// updateReservedQuantity updates the reserved quantity for a product
func (s *inventoryService) updateReservedQuantity(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, newReservedQuantity int) error {
	// Use UpdateReservedQuantity to properly persist reserved quantity
	// This fixes the bug where UpdateStock was called but only updated quantity, not reserved_quantity
	if err := s.repository.UpdateReservedQuantity(ctx, tenantID, productID, newReservedQuantity); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to update reserved quantity", err, map[string]interface{}{
			"product_id":            productID,
			"new_reserved_quantity": newReservedQuantity,
		})
		return common.CreateDatabaseError("update_reserved_quantity", err)
	}

	s.logger.DebugWithContext(ctx, "Updated reserved quantity", map[string]interface{}{
		"product_id":            productID,
		"new_reserved_quantity": newReservedQuantity,
	})
	return nil
}

// checkLowStockAlert checks if stock is below minimum levels and triggers alerts
func (s *inventoryService) checkLowStockAlert(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, availableStock int) {
	// Get stock record to check minimum levels
	stock, err := s.repository.GetStock(ctx, tenantID, productID)
	if err != nil {
		return
	}

	// Check if stock is below minimum level
	if availableStock <= stock.MinimumLevel {
		s.logger.WarnWithContext(ctx, "Low stock alert", map[string]interface{}{
			"product_id":      productID,
			"available_stock": availableStock,
			"minimum_level":   stock.MinimumLevel,
		})

		// Trigger low stock alert event
		common.PublishEvent(ctx, "inventory.low_stock", map[string]interface{}{
			"tenant_id":       tenantID,
			"product_id":      productID,
			"available_stock": availableStock,
			"minimum_level":   stock.MinimumLevel,
			"alert_type":      "low_stock",
		})
	}

	// Check if stock is at reorder point
	if availableStock <= stock.ReorderPoint {
		s.logger.InfoWithContext(ctx, "Reorder point reached", map[string]interface{}{
			"product_id":      productID,
			"available_stock": availableStock,
			"reorder_point":   stock.ReorderPoint,
		})

		// Trigger reorder point alert event
		common.PublishEvent(ctx, "inventory.reorder_point", map[string]interface{}{
			"tenant_id":       tenantID,
			"product_id":      productID,
			"available_stock": availableStock,
			"reorder_point":   stock.ReorderPoint,
			"alert_type":      "reorder",
		})
	}
} // Legacy methods implementation for InventoryService

// List retrieves all inventory records with pagination
func (s *inventoryService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*InventoryRecord, error) {
	inventories, err := s.repository.List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list inventories: %w", err)
	}
	return inventories, nil
}

// Create creates a new inventory record
func (s *inventoryService) Create(ctx context.Context, inventory *InventoryRecord) error {
	if err := s.repository.Create(ctx, inventory); err != nil {
		return fmt.Errorf("failed to create inventory: %w", err)
	}
	return nil
}

// GetByID retrieves an inventory record by ID
func (s *inventoryService) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*InventoryRecord, error) {
	inventory, err := s.repository.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	return inventory, nil
}

// Update updates an existing inventory record
func (s *inventoryService) Update(ctx context.Context, inventory *InventoryRecord) error {
	if err := s.repository.Update(ctx, inventory); err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}
	return nil
}

// Delete deletes an inventory record
func (s *inventoryService) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	if err := s.repository.Delete(ctx, tenantID, id); err != nil {
		return fmt.Errorf("failed to delete inventory: %w", err)
	}
	return nil
}

// BulkAdjustStock performs multiple stock adjustments
func (s *inventoryService) BulkAdjustStock(ctx context.Context, tenantID uuid.UUID, adjustments []repositories.BulkAdjustmentItem) error {
	if err := s.repository.BulkAdjust(ctx, tenantID, adjustments); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to bulk adjust stock", err, map[string]interface{}{
			"count": len(adjustments),
		})
		return fmt.Errorf("failed to bulk adjust stock: %w", err)
	}
	return nil
}

// BulkDelete deletes multiple inventory records
func (s *inventoryService) BulkDelete(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) error {
	if err := s.repository.BulkDelete(ctx, tenantID, ids); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to bulk delete inventory", err, map[string]interface{}{
			"count": len(ids),
		})
		return fmt.Errorf("failed to bulk delete inventory: %w", err)
	}
	return nil
}
