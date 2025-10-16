package jobs

import (
	"context"
	"log"

	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type InventoryAlertService struct {
	inventoryRepo repositories.InventoryRepository
	productRepo   repositories.ProductRepository
}

type InventoryAlert struct {
	TenantID     uuid.UUID
	WarehouseID  uuid.UUID
	ProductID    uuid.UUID
	ProductName  string
	CurrentStock int
	Threshold    int
}

func NewInventoryAlertService(inventoryRepo repositories.InventoryRepository, productRepo repositories.ProductRepository) *InventoryAlertService {
	return &InventoryAlertService{
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
	}
}

func (a *InventoryAlertService) CheckLowStock(ctx context.Context, tenantID uuid.UUID, threshold int) ([]InventoryAlert, error) {
	if threshold <= 0 {
		threshold = 10 // Default threshold
	}

	inventories, err := a.inventoryRepo.List(ctx, tenantID, 1000, 0) // Get all, in practice should paginate
	if err != nil {
		log.Printf("Failed to list inventories for tenant %s: %v", tenantID.String(), err)
		return nil, err
	}

	var alerts []InventoryAlert

	for _, inv := range inventories {
		if inv.Quantity <= threshold {
			// Get product name (this could be cached for performance)
			product, err := a.productRepo.GetByID(ctx, tenantID, inv.ProductID)
			if err != nil {
				log.Printf("Failed to get product %s: %v", inv.ProductID.String(), err)
				continue
			}

			alert := InventoryAlert{
				TenantID:     tenantID,
				WarehouseID:  inv.WarehouseID,
				ProductID:    inv.ProductID,
				ProductName:  product.Name,
				CurrentStock: inv.Quantity,
				Threshold:    threshold,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

func (a *InventoryAlertService) LogLowStockAlerts(ctx context.Context, alerts []InventoryAlert) {
	if len(alerts) == 0 {
		log.Println("No low stock alerts to log")
		return
	}

	log.Printf("Low stock alerts for tenant %s:", alerts[0].TenantID.String())
	for _, alert := range alerts {
		log.Printf("- Product '%s' in warehouse %s has %d units (threshold: %d)",
			alert.ProductName,
			alert.WarehouseID.String(),
			alert.CurrentStock,
			alert.Threshold)
	}
}

func (a *InventoryAlertService) CheckAndLogLowStockAcrossAllTenants(ctx context.Context, threshold int) error {
	// This would need a tenant repository to get all tenants
	// For now, we'll just log that this would runPeriodic

	log.Printf("Inventory alerts check would run with threshold: %d", threshold)
	log.Println("In production, this would check all tenants for low stock")

	return nil
}

// Scheduled job to run every hour
func (a *InventoryAlertService) ScheduledLowStockCheck(ctx context.Context) error {
	log.Println("Starting scheduled low stock check")

	err := a.CheckAndLogLowStockAcrossAllTenants(ctx, 10) // Default threshold of 10
	if err != nil {
		log.Printf("Scheduled low stock check failed: %v", err)
		return err
	}

	log.Println("Scheduled low stock check completed successfully")
	return nil
}