package jobs

import (
	"context"
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
	TenantsProcessed       int
	DataUpdated            bool
	MaterializedViewsRefed bool
	LastRefreshAt          time.Time
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

	result := &AnalyticsRefreshResult{
		TenantsProcessed:       0,
		DataUpdated:            false,
		MaterializedViewsRefed: false,
		LastRefreshAt:          time.Now(),
	}

	if a.db == nil {
		log.Println("Database pool not available, skipping tenant analytics refresh")
		return result, nil
	}

	// Query all active tenants
	query := `SELECT id FROM tenants WHERE status = 'active'`
	rows, err := a.db.Query(ctx, query)
	if err != nil {
		log.Printf("Failed to query tenants for analytics refresh: %v", err)
		return result, err
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

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating tenant rows: %v", err)
		return result, err
	}

	log.Printf("Found %d active tenants to refresh", len(tenantIDs))

	// Refresh analytics for each tenant
	successCount := 0
	failCount := 0
	for _, tenantID := range tenantIDs {
		if err := a.RefreshAnalyticsForTenant(ctx, tenantID); err != nil {
			log.Printf("Failed to refresh analytics for tenant %s: %v", tenantID.String(), err)
			failCount++
		} else {
			successCount++
		}
	}

	result.TenantsProcessed = successCount
	result.DataUpdated = successCount > 0

	log.Printf("Completed analytics refresh: processed=%d, succeeded=%d, failed=%d at %v",
		len(tenantIDs), successCount, failCount, result.LastRefreshAt.Format("2006-01-02 15:04:05"))

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

// RefreshMaterializedViews refreshes the PostgreSQL materialized views for analytics.
// This should be called periodically (e.g., every 15 minutes) to ensure dashboard
// analytics are pre-computed and queries are fast.
func (a *AnalyticsRefreshService) RefreshMaterializedViews(ctx context.Context) error {
	if a.db == nil {
		log.Println("Database pool not available, skipping materialized view refresh")
		return nil
	}

	log.Println("Refreshing materialized views for analytics...")
	startTime := time.Now()

	// Use CONCURRENTLY to avoid locking reads during refresh
	// This requires a UNIQUE index on the materialized view
	views := []string{
		"mv_dashboard_analytics",
		"mv_product_sales",
	}

	for _, view := range views {
		query := "REFRESH MATERIALIZED VIEW CONCURRENTLY " + view
		_, err := a.db.Exec(ctx, query)
		if err != nil {
			// If CONCURRENTLY fails (e.g., no unique index), try without it
			log.Printf("CONCURRENTLY refresh failed for %s, trying regular refresh: %v", view, err)
			query = "REFRESH MATERIALIZED VIEW " + view
			_, err = a.db.Exec(ctx, query)
			if err != nil {
				log.Printf("Failed to refresh materialized view %s: %v", view, err)
				// Continue with other views even if one fails
			}
		}
	}

	log.Printf("Materialized views refreshed in %v", time.Since(startTime))
	return nil
}

// Scheduled job for analytics refresh
func (a *AnalyticsRefreshService) ScheduledAnalyticsRefresh(ctx context.Context) error {
	log.Println("Running scheduled analytics refresh")

	StartTime := time.Now()
	defer func() {
		log.Printf("Scheduled analytics refresh completed in %v", time.Since(StartTime))
	}()

	// Refresh materialized views first for faster queries
	if err := a.RefreshMaterializedViews(ctx); err != nil {
		log.Printf("Materialized view refresh had errors: %v", err)
		// Continue with analytics refresh even if MV refresh fails
	}

	result, err := a.RefreshAllTenantsAnalytics(ctx)
	if err != nil {
		log.Printf("Scheduled analytics refresh failed: %v", err)
		return err
	}

	log.Printf("Successfully processed analytics for %d tenants", result.TenantsProcessed)
	return nil
}
