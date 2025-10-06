package analytics

import (
	"context"
	"fmt"
	"log"
	"time"

	"agromart2/internal/repositories"
	"agromart2/internal/caching"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AnalyticsService handles calculation and caching of analytics data
type AnalyticsService struct {
	orderRepo     repositories.OrderRepository
	invoiceRepo   repositories.InvoiceRepository
	inventoryRepo repositories.InventoryRepository
	productRepo   repositories.ProductRepository
	cacheService  caching.CacheService
	db            *pgxpool.Pool
}

// AnalyticsData represents cached analytics
type AnalyticsData struct {
	TenantID           uuid.UUID
	TotalSales         float64
	TotalStockValue    float64
	GSTCollected       float64
	OrderCount         int
	LowStockItemsCount int
	LastUpdated        time.Time
}

// SalesTrend represents sales data over time
type SalesTrend struct {
	Date     time.Time
	SalesAmount float64
	OrderCount  int
}

// SearchAnalytics represents search usage analytics
type SearchAnalytics struct {
	TenantID       uuid.UUID
	EntityType     string // "products", "orders", "inventory"
	SearchTerm     string
	FilterCount    int
	ResultCount    int
	Timestamp      time.Time
	UserID         uuid.UUID
	ResponseTimeMs int64
}

// SearchUsageStats represents aggregated search usage statistics
type SearchUsageStats struct {
	TenantID         uuid.UUID
	TotalSearches    int
	AvgResponseTime  float64
	TopSearchTerms   []SearchTermFrequency
	PeakUsageTimes   []TimeUsage
	DateRange        struct {
		Start time.Time
		End   time.Time
	}
}

// SearchTermFrequency represents frequency of search terms
type SearchTermFrequency struct {
	Term      string
	Frequency int
}

// TimeUsage represents usage by time period
type TimeUsage struct {
	Hour  int
	Count int
}

func NewAnalyticsService(orderRepo repositories.OrderRepository, invoiceRepo repositories.InvoiceRepository, inventoryRepo repositories.InventoryRepository, productRepo repositories.ProductRepository, cacheService caching.CacheService, db *pgxpool.Pool) *AnalyticsService {
	return &AnalyticsService{
		orderRepo:     orderRepo,
		invoiceRepo:   invoiceRepo,
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
		cacheService:  cacheService,
		db:            db,
	}
}

func (a *AnalyticsService) CalculateTenantAnalytics(ctx context.Context, tenantID uuid.UUID) (*AnalyticsData, error) {
	data := &AnalyticsData{
		TenantID:    tenantID,
		LastUpdated: time.Now(),
	}

	// Calculate total sales from invoices
	invoices, err := a.invoiceRepo.List(ctx, tenantID, 10000, 0) // Get all, should paginate in production
	if err != nil {
		log.Printf("Failed to get invoices for analytics: %v", err)
		return data, err
	}

	var totalSales float64
	var gstCollected float64
	for _, invoice := range invoices {
		totalSales += invoice.TotalAmount
		if invoice.CGST != nil {
			gstCollected += *invoice.CGST
		}
		if invoice.IGST != nil {
			gstCollected += *invoice.IGST
		}
		if invoice.SGST != nil {
			gstCollected += *invoice.SGST
		}
	}

	data.TotalSales = totalSales
	data.GSTCollected = gstCollected
	data.OrderCount = len(invoices)

	// Calculate total stock value
	inventories, err := a.inventoryRepo.List(ctx, tenantID, 10000, 0) // Get all
	if err != nil {
		log.Printf("Failed to get inventories for analytics: %v", err)
		return data, err
	}

	var totalStockValue float64
	lowStockCount := 0
	for _, inv := range inventories {
		if inv.Quantity < 10 { // Low stock threshold
			lowStockCount++
		}

		// Get product price to calculate stock value
		product, err := a.productRepo.GetByID(ctx, tenantID, inv.ProductID)
		if err != nil {
			log.Printf("Failed to get product %s: %v", inv.ProductID.String(), err)
			continue
		}
		totalStockValue += float64(inv.Quantity) * product.UnitPrice
	}

	data.TotalStockValue = totalStockValue
	data.LowStockItemsCount = lowStockCount

	return data, nil
}

func (a *AnalyticsService) GetSalesTrends(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]SalesTrend, error) {
	orders, err := a.orderRepo.GetOrdersByTenantAndDateRange(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Group by date
	trends := make(map[string]*SalesTrend)

	for _, order := range orders {
		dateStr := order.OrderDate.Format("2006-01-02")
		if trends[dateStr] == nil {
			trends[dateStr] = &SalesTrend{
				Date: order.OrderDate,
			}
		}
		trends[dateStr].SalesAmount += float64(order.Quantity) * order.UnitPrice
		trends[dateStr].OrderCount++
	}

	var result []SalesTrend
	for _, trend := range trends {
		result = append(result, *trend)
	}

	return result, nil
}

func (a *AnalyticsService) CalculateGSTTotals(ctx context.Context, tenantID uuid.UUID) (map[string]float64, error) {
	invoices, err := a.invoiceRepo.List(ctx, tenantID, 10000, 0)
	if err != nil {
		return nil, err
	}

	totals := map[string]float64{
		"cgst": 0,
		"sgst": 0,
		"igst": 0,
		"total": 0,
	}

	for _, invoice := range invoices {
		if invoice.CGST != nil {
			totals["cgst"] += *invoice.CGST
			totals["total"] += *invoice.CGST
		}
		if invoice.SGST != nil {
			totals["sgst"] += *invoice.SGST
			totals["total"] += *invoice.SGST
		}
		if invoice.IGST != nil {
			totals["igst"] += *invoice.IGST
			totals["total"] += *invoice.IGST
		}
	}

	return totals, nil
}

// RecordSearchUsage tracks search operations for analytics by persisting to database
func (a *AnalyticsService) RecordSearchUsage(ctx context.Context, tenantID uuid.UUID, entityType string, searchTerm string, filterCount int, resultCount int, userID uuid.UUID, responseTimeMs int64) error {
	query := `
		INSERT INTO search_analytics (tenant_id, entity_type, search_term, filter_count, result_count, user_id, response_time_ms, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	
	_, err := a.db.Exec(ctx, query, 
		tenantID, entityType, searchTerm, filterCount, resultCount, userID, responseTimeMs)
	
	if err != nil {
		log.Printf("Failed to persist search analytics: %v", err)
		// Don't fail the operation, just log the error
		return nil
	}

	log.Printf("Search Usage - Tenant: %s, Entity: %s, Term: '%s', Filters: %d, Results: %d, User: %s, Response: %dms",
		tenantID.String(), entityType, searchTerm, filterCount, resultCount, userID.String(), responseTimeMs)

	return nil
}

// GetSearchAnalytics retrieves search usage statistics for a tenant from database
func (a *AnalyticsService) GetSearchAnalytics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*SearchUsageStats, error) {
	stats := &SearchUsageStats{
		TenantID: tenantID,
		DateRange: struct {
			Start time.Time
			End   time.Time
		}{
			Start: startDate,
			End:   endDate,
		},
	}

	// Query total searches and average response time
	countQuery := `
		SELECT COUNT(*), COALESCE(AVG(response_time_ms), 0)
		FROM search_analytics
		WHERE tenant_id = $1 AND timestamp BETWEEN $2 AND $3
	`
	
	err := a.db.QueryRow(ctx, countQuery, tenantID, startDate, endDate).
		Scan(&stats.TotalSearches, &stats.AvgResponseTime)
	if err != nil {
		log.Printf("Failed to get search count: %v", err)
		return stats, nil // Return empty stats on error
	}

	// Query top search terms
	termsQuery := `
		SELECT search_term, COUNT(*) as frequency
		FROM search_analytics
		WHERE tenant_id = $1 AND timestamp BETWEEN $2 AND $3
		  AND search_term IS NOT NULL AND search_term != ''
		GROUP BY search_term
		ORDER BY frequency DESC
		LIMIT 10
	`
	
	rows, err := a.db.Query(ctx, termsQuery, tenantID, startDate, endDate)
	if err != nil {
		log.Printf("Failed to get top search terms: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var term SearchTermFrequency
			if err := rows.Scan(&term.Term, &term.Frequency); err == nil {
				stats.TopSearchTerms = append(stats.TopSearchTerms, term)
			}
		}
	}

	// Query peak usage times (by hour)
	usageQuery := `
		SELECT EXTRACT(HOUR FROM timestamp)::int as hour, COUNT(*) as count
		FROM search_analytics
		WHERE tenant_id = $1 AND timestamp BETWEEN $2 AND $3
		GROUP BY hour
		ORDER BY count DESC
		LIMIT 5
	`
	
	usageRows, err := a.db.Query(ctx, usageQuery, tenantID, startDate, endDate)
	if err != nil {
		log.Printf("Failed to get peak usage times: %v", err)
	} else {
		defer usageRows.Close()
		for usageRows.Next() {
			var usage TimeUsage
			if err := usageRows.Scan(&usage.Hour, &usage.Count); err == nil {
				stats.PeakUsageTimes = append(stats.PeakUsageTimes, usage)
			}
		}
	}

	return stats, nil
}

// TrackBulkOperationUsage tracks bulk operation usage by persisting to database
func (a *AnalyticsService) TrackBulkOperationUsage(ctx context.Context, tenantID uuid.UUID, operationType string, totalItems int, successCount int, userID uuid.UUID, processingTimeMs int64) error {
	failureCount := totalItems - successCount
	
	query := `
		INSERT INTO bulk_operation_analytics (tenant_id, operation_type, total_items, success_count, failure_count, user_id, processing_time_ms, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	
	_, err := a.db.Exec(ctx, query,
		tenantID, operationType, totalItems, successCount, failureCount, userID, processingTimeMs)
	
	if err != nil {
		log.Printf("Failed to persist bulk operation analytics: %v", err)
		// Don't fail the operation, just log the error
		return nil
	}

	log.Printf("Bulk Operation Usage - Tenant: %s, Type: %s, Total: %d, Success: %d, User: %s, Processing: %dms",
		tenantID.String(), operationType, totalItems, successCount, userID.String(), processingTimeMs)

	return nil
}

// GetPopularSearchTerms returns the most frequently searched terms from database
func (a *AnalyticsService) GetPopularSearchTerms(ctx context.Context, tenantID uuid.UUID, limit int) ([]SearchTermFrequency, error) {
	if limit <= 0 {
		limit = 10
	}
	
	query := `
		SELECT search_term, COUNT(*) as frequency
		FROM search_analytics
		WHERE tenant_id = $1
		  AND search_term IS NOT NULL AND search_term != ''
		  AND timestamp > NOW() - INTERVAL '30 days'
		GROUP BY search_term
		ORDER BY frequency DESC
		LIMIT $2
	`
	
	rows, err := a.db.Query(ctx, query, tenantID, limit)
	if err != nil {
		log.Printf("Failed to get popular search terms: %v", err)
		return []SearchTermFrequency{}, nil
	}
	defer rows.Close()

	var terms []SearchTermFrequency
	for rows.Next() {
		var term SearchTermFrequency
		if err := rows.Scan(&term.Term, &term.Frequency); err == nil {
			terms = append(terms, term)
		}
	}

	return terms, nil
}

// GetStockLevels returns current stock levels for a tenant's products
func (a *AnalyticsService) GetStockLevels(ctx context.Context, tenantID uuid.UUID) ([]struct {
	ProductName string
	Quantity    int
}, error) {
	inventories, err := a.inventoryRepo.List(ctx, tenantID, 10000, 0)
	if err != nil {
		return nil, err
	}

	var stockLevels []struct {
		ProductName string
		Quantity    int
	}

	for _, inv := range inventories {
		product, err := a.productRepo.GetByID(ctx, tenantID, inv.ProductID)
		if err != nil {
			continue // Skip if product not found
		}

		stockLevels = append(stockLevels, struct {
			ProductName string
			Quantity    int
		}{
			ProductName: product.Name,
			Quantity:    inv.Quantity,
		})
	}

	return stockLevels, nil
}

// GetSearchPerformanceMetrics returns search performance metrics from database
func (a *AnalyticsService) GetSearchPerformanceMetrics(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Query basic metrics
	basicQuery := `
		SELECT 
			COALESCE(AVG(response_time_ms), 0) as avg_response_time,
			COUNT(*) as total_searches,
			SUM(CASE WHEN result_count > 0 THEN 1 ELSE 0 END) as successful_searches,
			SUM(CASE WHEN result_count = 0 THEN 1 ELSE 0 END) as zero_result_searches
		FROM search_analytics
		WHERE tenant_id = $1 AND timestamp > NOW() - INTERVAL '30 days'
	`
	
	var avgResponseTime float64
	var totalSearches, successfulSearches, zeroResultSearches int64
	
	err := a.db.QueryRow(ctx, basicQuery, tenantID).
		Scan(&avgResponseTime, &totalSearches, &successfulSearches, &zeroResultSearches)
	if err != nil {
		log.Printf("Failed to get search performance metrics: %v", err)
		return metrics, nil
	}

	metrics["avg_response_time_ms"] = avgResponseTime
	metrics["total_searches"] = totalSearches
	metrics["successful_searches"] = successfulSearches
	metrics["failed_searches"] = totalSearches - successfulSearches
	
	var zeroResultPct float64
	if totalSearches > 0 {
		zeroResultPct = float64(zeroResultSearches) / float64(totalSearches) * 100
	}
	metrics["zero_result_searches_pct"] = zeroResultPct

	// Get most popular entity type
	entityQuery := `
		SELECT entity_type, COUNT(*) as count
		FROM search_analytics
		WHERE tenant_id = $1 AND timestamp > NOW() - INTERVAL '30 days'
		GROUP BY entity_type
		ORDER BY count DESC
		LIMIT 1
	`
	
	var mostPopularEntity string
	var entityCount int64
	err = a.db.QueryRow(ctx, entityQuery, tenantID).
		Scan(&mostPopularEntity, &entityCount)
	if err == nil {
		metrics["most_popular_entity"] = mostPopularEntity
	} else {
		metrics["most_popular_entity"] = "products"
	}

	// Get peak usage hour
	peakQuery := `
		SELECT EXTRACT(HOUR FROM timestamp)::int as hour, COUNT(*) as count
		FROM search_analytics
		WHERE tenant_id = $1 AND timestamp > NOW() - INTERVAL '30 days'
		GROUP BY hour
		ORDER BY count DESC
		LIMIT 1
	`
	
	var peakHour int
	var peakCount int64
	err = a.db.QueryRow(ctx, peakQuery, tenantID).
		Scan(&peakHour, &peakCount)
	if err == nil {
		metrics["peak_usage_hour"] = peakHour
	} else {
		metrics["peak_usage_hour"] = 14 // Default to 2 PM
	}

	// Placeholder for most used filters (would require additional tracking)
	metrics["most_used_filters"] = []string{"category", "quantity", "price_range"}

	return metrics, nil
}

// InvalidateTenantAnalyticsCache invalidates cached analytics data for a tenant
func (a *AnalyticsService) InvalidateTenantAnalyticsCache(ctx context.Context, tenantID uuid.UUID) error {
	log.Printf("Invalidating analytics cache for tenant %s", tenantID.String())

	// Use cache service to invalidate by pattern or specific key
	// Since we want to invalidate only analytics, we can delete the specific key
	cacheKey := fmt.Sprintf("agromart:analytics:%s", tenantID.String())
	return a.cacheService.Delete(ctx, cacheKey)
}
