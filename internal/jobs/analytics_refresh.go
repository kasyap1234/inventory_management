package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"agromart2/internal/analytics"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRefreshService struct {
	analyticsService *analytics.AnalyticsService
	db               *pgxpool.Pool
}

type AnalyticsRefreshResult struct {
	TenantsProcessed int
	DataUpdated      bool
	LastRefreshAt    time.Time
}

func NewAnalyticsRefreshService(analyticsService *analytics.AnalyticsService, db *pgxpool.Pool) *AnalyticsRefreshService {
	return &AnalyticsRefreshService{
		analyticsService: analyticsService,
		db:               db,
	}
}

func (a *AnalyticsRefreshService) RefreshAnalyticsForTenant(ctx context.Context, tenantID uuid.UUID) error {
	log.Printf("Refreshing analytics for tenant: %s", tenantID.String())

	data, err := a.analyticsService.CalculateTenantAnalytics(ctx, tenantID)
	if err != nil {
		log.Printf("Failed to calculate analytics for tenant %s: %v", tenantID.String(), err)
		return err
	}

	// In a real implementation, this would save to Redis or database
	log.Printf("Analytics updated for tenant %s: Sales=%.2f, StockValue=%.2f, GSTCollected=%.2f, LowStockItems=%d",
		tenantID.String(), data.TotalSales, data.TotalStockValue, data.GSTCollected, data.LowStockItemsCount)

	return nil
}

func (a *AnalyticsRefreshService) RefreshAllTenantsAnalytics(ctx context.Context) (*AnalyticsRefreshResult, error) {
	log.Println("Starting analytics refresh for all tenants")

	// In a real implementation, this would get all tenant IDs from tenant repository
	// For now, just simulate processing

	// Query all active tenants from database
	query := `SELECT id FROM tenants WHERE deleted_at IS NULL ORDER BY created_at`
	rows, err := a.db.Query(ctx, query)
	if err != nil {
		log.Printf("Failed to fetch tenants for analytics refresh: %v", err)
		return nil, fmt.Errorf("failed to fetch tenants: %w", err)
	}
	defer rows.Close()

	var tenantIDs []uuid.UUID
	for rows.Next() {
		var tenantID uuid.UUID
		if err := rows.Scan(&tenantID); err != nil {
			log.Printf("Failed to scan tenant ID: %v", err)
			continue
		}
		tenantIDs = append(tenantIDs, tenantID)
	}

	// Process analytics refresh for each tenant
	successCount := 0
	for _, tenantID := range tenantIDs {
		log.Printf("Refreshing analytics for tenant: %s", tenantID.String())
		
		// Refresh analytics for this tenant
		if err := a.RefreshStockLevelsDashboard(ctx, tenantID); err != nil {
			log.Printf("Failed to refresh analytics for tenant %s: %v", tenantID, err)
			continue
		}
		
		// Invalidate cache for this tenant
		if err := a.analyticsService.InvalidateTenantAnalyticsCache(ctx, tenantID); err != nil {
			log.Printf("Failed to invalidate cache for tenant %s: %v", tenantID, err)
		}
		
		successCount++
	}

	result := &AnalyticsRefreshResult{
		TenantsProcessed: successCount,
		DataUpdated:      true,
		LastRefreshAt:    time.Now(),
	}

	log.Printf("Completed analytics refresh for %d/%d tenants at %v",
		result.TenantsProcessed, len(tenantIDs), result.LastRefreshAt.Format("2006-01-02 15:04:05"))

	return result, nil
}

func (a *AnalyticsRefreshService) RefreshStockLevelsDashboard(ctx context.Context, tenantID uuid.UUID) error {
	log.Printf("Refreshing stock levels dashboard for tenant: %s", tenantID.String())

	inventory, err := a.analyticsService.GetStockLevels(ctx, tenantID)
	if err != nil {
		log.Printf("Failed to refresh stock levels: %v", err)
		return err
	}

	// Log stock levels (in Redis/database this would be cached)
	if len(inventory) > 0 {
		for _, item := range inventory {
			log.Printf("Stock: Product %s - Qty: %d", item.ProductName, item.Quantity)
		}
	}

	return nil
}

// Scheduled job for analytics refresh
func (a *AnalyticsRefreshService) ScheduledAnalyticsRefresh(ctx context.Context) error {
	log.Println("Running scheduled analytics refresh")

	StartTime := time.Now()
	defer func() {
		log.Printf("Scheduled analytics refresh completed in %v", time.Since(StartTime))
	}()

	result, err := a.RefreshAllTenantsAnalytics(ctx)
	if err != nil {
		log.Printf("Scheduled analytics refresh failed: %v", err)
		return err
	}

	log.Printf("Successfully processed analytics for %d tenants", result.TenantsProcessed)
	return nil
}