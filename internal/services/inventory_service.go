package services

import (
	"context"
	"fmt"
	"time"

	"agromart2/internal/caching"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type InventoryService interface {
	Create(ctx context.Context, tenantID uuid.UUID, inventory *models.Inventory) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Inventory, error)
	Update(ctx context.Context, tenantID uuid.UUID, inventory *models.Inventory) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Inventory, error)
	Transfer(ctx context.Context, tenantID, productID, fromWarehouseID, toWarehouseID uuid.UUID, quantity int) error
	AdjustStock(ctx context.Context, tenantID, warehouseID, productID uuid.UUID, quantityChange int) error
	LowStockAlerts(ctx context.Context, tenantID uuid.UUID, threshold int) ([]*models.Inventory, error)
	GetByWarehouseAndProduct(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) (*models.Inventory, error)
	AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.InventorySearchFilter) ([]*models.Inventory, error)

	// Bulk operations
	BulkAdjustStock(ctx context.Context, tenantID uuid.UUID, bulkAdjust *models.InventoryBulkAdjust) (*models.BulkOperationResult, error)
	BulkTransferStock(ctx context.Context, tenantID uuid.UUID, bulkTransfer *models.InventoryBulkTransfer) (*models.BulkOperationResult, error)
}

type inventoryService struct {
	inventoryRepo repositories.InventoryRepository
	productRepo   repositories.ProductRepository
	cacheService  caching.CacheService
}

func NewInventoryService(inventoryRepo repositories.InventoryRepository, productRepo repositories.ProductRepository, cacheService caching.CacheService) InventoryService {
	return &inventoryService{
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
		cacheService:  cacheService,
	}
}

func (s *inventoryService) Create(ctx context.Context, tenantID uuid.UUID, inventory *models.Inventory) error {
	inventory.TenantID = tenantID
	inventory.ID = uuid.New()
	return s.inventoryRepo.Create(ctx, inventory)
}

func (s *inventoryService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Inventory, error) {
	return s.inventoryRepo.GetByID(ctx, tenantID, id)
}

func (s *inventoryService) Update(ctx context.Context, tenantID uuid.UUID, inventory *models.Inventory) error {
	inventory.TenantID = tenantID

	err := s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return err
	}

	// Invalidate cache for this inventory
	if cacheErr := s.cacheService.DeleteInventory(ctx, tenantID, inventory.WarehouseID, inventory.ProductID); cacheErr != nil {
		fmt.Printf("Failed to invalidate cache for inventory %s-%s: %v\n", inventory.WarehouseID.String(), inventory.ProductID.String(), cacheErr)
	}

	return nil
}

func (s *inventoryService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.inventoryRepo.Delete(ctx, tenantID, id)
}

func (s *inventoryService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Inventory, error) {
	return s.inventoryRepo.List(ctx, tenantID, limit, offset)
}

func (s *inventoryService) Transfer(ctx context.Context, tenantID, productID, fromWarehouseID, toWarehouseID uuid.UUID, quantity int) error {
	fromInventory, err := s.inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, fromWarehouseID, productID)
	if err != nil {
		return err
	}
	if fromInventory.Quantity < quantity {
		return err // Insufficient stock
	}
	toInventory, err := s.inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, toWarehouseID, productID)
	if err != nil {
		// Create new inventory if not exists
		toInventory = &models.Inventory{
			TenantID:    tenantID,
			WarehouseID: toWarehouseID,
			ProductID:   productID,
			Quantity:    0,
		}
		toInventory.ID = uuid.New()
		s.inventoryRepo.Create(ctx, toInventory)
	}
	fromInventory.Quantity -= quantity
	toInventory.Quantity += quantity

	if err := s.inventoryRepo.Update(ctx, fromInventory); err != nil {
		return err
	}
	if err := s.inventoryRepo.Update(ctx, toInventory); err != nil {
		return err
	}

	// Invalidate caches for both inventories
	if cacheErr := s.cacheService.DeleteInventory(ctx, tenantID, fromWarehouseID, productID); cacheErr != nil {
		fmt.Printf("Failed to invalidate cache for source inventory %s-%s: %v\n", fromWarehouseID.String(), productID.String(), cacheErr)
	}
	if cacheErr := s.cacheService.DeleteInventory(ctx, tenantID, toWarehouseID, productID); cacheErr != nil {
		fmt.Printf("Failed to invalidate cache for destination inventory %s-%s: %v\n", toWarehouseID.String(), productID.String(), cacheErr)
	}

	return nil
}

func (s *inventoryService) AdjustStock(ctx context.Context, tenantID, warehouseID, productID uuid.UUID, quantityChange int) error {
	inventory, err := s.inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, warehouseID, productID)
	if err != nil {
		// Assume warehouse and product exist
		inventory = &models.Inventory{
			TenantID:    tenantID,
			WarehouseID: warehouseID,
			ProductID:   productID,
			Quantity:    0,
		}
		inventory.ID = uuid.New()
		err := s.inventoryRepo.Create(ctx, inventory)
		if err != nil {
			return err
		}
	}
	inventory.Quantity += quantityChange
	if inventory.Quantity < 0 {
		inventory.Quantity = 0 // Prevent negative
	}

	err = s.inventoryRepo.Update(ctx, inventory)
	if err != nil {
		return err
	}

	// Invalidate cache for this inventory
	if cacheErr := s.cacheService.DeleteInventory(ctx, tenantID, warehouseID, productID); cacheErr != nil {
		fmt.Printf("Failed to invalidate cache for adjusted inventory %s-%s: %v\n", warehouseID.String(), productID.String(), cacheErr)
	}

	return nil
}

func (s *inventoryService) LowStockAlerts(ctx context.Context, tenantID uuid.UUID, threshold int) ([]*models.Inventory, error) {
	all, err := s.inventoryRepo.List(ctx, tenantID, 1000, 0) // Simplified
	if err != nil {
		return nil, err
	}
	var low []*models.Inventory
	for _, inv := range all {
		if inv.Quantity <= threshold {
			low = append(low, inv)
		}
	}
	return low, nil
}

func (s *inventoryService) GetByWarehouseAndProduct(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) (*models.Inventory, error) {
	// Try to get from cache first
	if cachedInventory, err := s.cacheService.GetInventory(ctx, tenantID, warehouseID, productID); cachedInventory != nil {
		return cachedInventory, nil
	} else if err != nil {
		// Log error but continue to database - cache errors shouldn't fail the operation
		fmt.Printf("Cache error for inventory %s-%s: %v\n", warehouseID.String(), productID.String(), err)
	}

	// Cache miss - get from database
	inventory, err := s.inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, warehouseID, productID)
	if err != nil {
		return nil, err
	}

	// Cache the inventory for future requests (TTL: 5 minutes since inventory changes frequently)
	if cacheErr := s.cacheService.SetInventory(ctx, tenantID, inventory, 5*time.Minute); cacheErr != nil {
		fmt.Printf("Failed to cache inventory %s-%s: %v\n", warehouseID.String(), productID.String(), cacheErr)
	}

	return inventory, nil
}

func (s *inventoryService) AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.InventorySearchFilter) ([]*models.Inventory, error) {
	return s.inventoryRepo.AdvancedSearch(ctx, tenantID, filter)
}

// BulkAdjustStock performs bulk stock adjustments
func (s *inventoryService) BulkAdjustStock(ctx context.Context, tenantID uuid.UUID, bulkAdjust *models.InventoryBulkAdjust) (*models.BulkOperationResult, error) {
	// Set defaults
	if bulkAdjust.ValidationMode == "" {
		bulkAdjust.ValidationMode = "strict"
	}
	if bulkAdjust.TransactionMode == "" {
		bulkAdjust.TransactionMode = "atomic"
	}

	result := &models.BulkOperationResult{
		OperationID:   fmt.Sprintf("bulk_adjust_stock_%d", time.Now().UnixNano()),
		Status:       "processing",
		TotalItems:    len(bulkAdjust.Adjustments),
		StartTime:     time.Now(),
		Progress:      0,
		Errors:        []models.BulkOperationError{},
		Items:         []models.BulkOperationItem{},
	}

	totalItems := len(bulkAdjust.Adjustments)

	for i, adjustment := range bulkAdjust.Adjustments {
		// Get existing inventory or create new one
		inventory, err := s.inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, adjustment.WarehouseID, adjustment.ProductID)
		if err != nil {
			// Create new inventory record if it doesn't exist
			inventory = &models.Inventory{
				TenantID:    tenantID,
				WarehouseID: adjustment.WarehouseID,
				ProductID:   adjustment.ProductID,
				Quantity:    0,
			}
			inventory.ID = uuid.New()
			err = s.inventoryRepo.Create(ctx, inventory)
			if err != nil {
				result.FailedItems++
				errorMsg := fmt.Sprintf("Failed to create inventory record: %v", err)
				result.Errors = append(result.Errors, models.BulkOperationError{
					ItemIndex: i,
					ItemID:    fmt.Sprintf("%s-%s", adjustment.WarehouseID.String(), adjustment.ProductID.String()),
					Error:     errorMsg,
				})
				result.Items = append(result.Items, models.BulkOperationItem{
					ItemIndex: i,
					ItemID:    fmt.Sprintf("%s-%s", adjustment.WarehouseID.String(), adjustment.ProductID.String()),
					Status:    "failed",
					Error:     &errorMsg,
				})
				continue
			}
		}

		// Check if we have enough stock for deduction
		if adjustment.QuantityChange < 0 && inventory.Quantity < -adjustment.QuantityChange {
			if bulkAdjust.ValidationMode == "strict" {
				result.FailedItems++
				errorMsg := fmt.Sprintf("Insufficient stock: available %d, requested deduction of %d",
					inventory.Quantity, -adjustment.QuantityChange)
				result.Errors = append(result.Errors, models.BulkOperationError{
					ItemIndex: i,
					ItemID:    fmt.Sprintf("%s-%s", adjustment.WarehouseID.String(), adjustment.ProductID.String()),
					Error:     errorMsg,
				})
				result.Items = append(result.Items, models.BulkOperationItem{
					ItemIndex: i,
					ItemID:    fmt.Sprintf("%s-%s", adjustment.WarehouseID.String(), adjustment.ProductID.String()),
					Status:    "failed",
					Error:     &errorMsg,
				})
				continue
			}
			// For skip_invalid mode, set to 0 instead of going negative
			inventory.Quantity = 0
		} else {
			inventory.Quantity += adjustment.QuantityChange
			if inventory.Quantity < 0 {
				inventory.Quantity = 0
			}
		}

		// Update inventory
		err = s.inventoryRepo.Update(ctx, inventory)
		if err != nil {
			result.FailedItems++
			errorMsg := fmt.Sprintf("Failed to update inventory: %v", err)
			result.Errors = append(result.Errors, models.BulkOperationError{
				ItemIndex: i,
				ItemID:    fmt.Sprintf("%s-%s", adjustment.WarehouseID.String(), adjustment.ProductID.String()),
				Error:     errorMsg,
			})
			result.Items = append(result.Items, models.BulkOperationItem{
				ItemIndex: i,
				ItemID:    fmt.Sprintf("%s-%s", adjustment.WarehouseID.String(), adjustment.ProductID.String()),
				Status:    "failed",
				Error:     &errorMsg,
			})
		} else {
			result.ProcessedItems++
			result.Items = append(result.Items, models.BulkOperationItem{
				ItemIndex: i,
				ItemID:    fmt.Sprintf("%s-%s", adjustment.WarehouseID.String(), adjustment.ProductID.String()),
				Status:    "success",
			})
		}

		// Update progress
		result.Progress = float64(i+1) / float64(totalItems) * 100
	}

	result.Status = "completed"
	if result.FailedItems > 0 && result.ProcessedItems > 0 {
		result.Status = "partial"
	}
	result.CompletionTime = &time.Time{}
	*result.CompletionTime = time.Now()

	return result, nil
}

// BulkTransferStock performs bulk stock transfers between warehouses
func (s *inventoryService) BulkTransferStock(ctx context.Context, tenantID uuid.UUID, bulkTransfer *models.InventoryBulkTransfer) (*models.BulkOperationResult, error) {
	// Set defaults
	if bulkTransfer.ValidationMode == "" {
		bulkTransfer.ValidationMode = "strict"
	}
	if bulkTransfer.TransactionMode == "" {
		bulkTransfer.TransactionMode = "atomic"
	}

	result := &models.BulkOperationResult{
		OperationID:   fmt.Sprintf("bulk_transfer_stock_%d", time.Now().UnixNano()),
		Status:       "processing",
		TotalItems:    len(bulkTransfer.Transfers),
		StartTime:     time.Now(),
		Progress:      0,
		Errors:        []models.BulkOperationError{},
		Items:         []models.BulkOperationItem{},
	}

	totalItems := len(bulkTransfer.Transfers)

	for i, transfer := range bulkTransfer.Transfers {
		// Check if source warehouse has enough stock
		fromInventory, err := s.inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, transfer.FromWarehouseID, transfer.ProductID)
		if err != nil {
			result.FailedItems++
			errorMsg := fmt.Sprintf("Source inventory not found: %v", err)
			result.Errors = append(result.Errors, models.BulkOperationError{
				ItemIndex: i,
				ItemID:    fmt.Sprintf("%s-%s", transfer.FromWarehouseID.String(), transfer.ProductID.String()),
				Error:     errorMsg,
			})
			result.Items = append(result.Items, models.BulkOperationItem{
				ItemIndex: i,
				ItemID:    fmt.Sprintf("%s-%s", transfer.FromWarehouseID.String(), transfer.ProductID.String()),
				Status:    "failed",
				Error:     &errorMsg,
			})
			continue
		}

		if fromInventory.Quantity < transfer.Quantity {
			result.FailedItems++
			errorMsg := fmt.Sprintf("Insufficient stock in source warehouse: available %d, requested %d",
				fromInventory.Quantity, transfer.Quantity)
			result.Errors = append(result.Errors, models.BulkOperationError{
				ItemIndex: i,
				ItemID:    fmt.Sprintf("%s-%s", transfer.FromWarehouseID.String(), transfer.ProductID.String()),
				Error:     errorMsg,
			})
			result.Items = append(result.Items, models.BulkOperationItem{
				ItemIndex: i,
				ItemID:    fmt.Sprintf("%s-%s", transfer.FromWarehouseID.String(), transfer.ProductID.String()),
				Status:    "failed",
				Error:     &errorMsg,
			})
			continue
		}

		// Get or create destination inventory
		toInventory, err := s.inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, transfer.ToWarehouseID, transfer.ProductID)
		if err != nil {
			// Create new inventory record
			toInventory = &models.Inventory{
				TenantID:    tenantID,
				WarehouseID: transfer.ToWarehouseID,
				ProductID:   transfer.ProductID,
				Quantity:    0,
			}
			toInventory.ID = uuid.New()
			err = s.inventoryRepo.Create(ctx, toInventory)
			if err != nil {
				result.FailedItems++
				errorMsg := "Failed to create destination inventory record"
				result.Errors = append(result.Errors, models.BulkOperationError{
					ItemIndex: i,
					ItemID:    fmt.Sprintf("%s-%s", transfer.ToWarehouseID.String(), transfer.ProductID.String()),
					Error:     errorMsg,
				})
				result.Items = append(result.Items, models.BulkOperationItem{
					ItemIndex: i,
					ItemID:    fmt.Sprintf("%s-%s", transfer.ToWarehouseID.String(), transfer.ProductID.String()),
					Status:    "failed",
					Error:     &errorMsg,
				})
				continue
			}
		}

		// Perform transfer
		fromInventory.Quantity -= transfer.Quantity
		toInventory.Quantity += transfer.Quantity

		// Update both records
		err1 := s.inventoryRepo.Update(ctx, fromInventory)
		err2 := s.inventoryRepo.Update(ctx, toInventory)

		if err1 != nil || err2 != nil {
			result.FailedItems++
			errorMsg := "Failed to update inventory records during transfer"
			if err1 != nil {
				errorMsg += fmt.Sprintf(" (source: %v)", err1)
			}
			if err2 != nil {
				errorMsg += fmt.Sprintf(" (destination: %v)", err2)
			}
			result.Errors = append(result.Errors, models.BulkOperationError{
				ItemIndex: i,
				ItemID:    fmt.Sprintf("%s-%s-%s", transfer.ProductID.String(), transfer.FromWarehouseID.String(), transfer.ToWarehouseID.String()),
				Error:     errorMsg,
			})
			result.Items = append(result.Items, models.BulkOperationItem{
				ItemIndex: i,
				ItemID:    fmt.Sprintf("%s-%s-%s", transfer.ProductID.String(), transfer.FromWarehouseID.String(), transfer.ToWarehouseID.String()),
				Status:    "failed",
				Error:     &errorMsg,
			})
		} else {
			result.ProcessedItems++
			result.Items = append(result.Items, models.BulkOperationItem{
				ItemIndex: i,
				ItemID:    fmt.Sprintf("%s-%s-%s", transfer.ProductID.String(), transfer.FromWarehouseID.String(), transfer.ToWarehouseID.String()),
				Status:    "success",
			})
		}

		// Update progress
		result.Progress = float64(i+1) / float64(totalItems) * 100
	}

	result.Status = "completed"
	if result.FailedItems > 0 && result.ProcessedItems > 0 {
		result.Status = "partial"
	}
	result.CompletionTime = &time.Time{}
	*result.CompletionTime = time.Now()

	return result, nil
}