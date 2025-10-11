package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"agromart2/internal/caching"
	"agromart2/internal/common"
	"agromart2/internal/repositories"

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
	Date        time.Time
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
	TenantID        uuid.UUID
	TotalSearches   int
	AvgResponseTime float64
	TopSearchTerms  []SearchTermFrequency
	PeakUsageTimes  []TimeUsage
	DateRange       struct {
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

const (
	defaultAnalyticsCacheTTL        = 15 * time.Minute
	salesTrendCacheTTL              = 10 * time.Minute
	gstTotalsCacheTTL               = 15 * time.Minute
	productSalesCacheTTL            = 15 * time.Minute
	lowStockCacheTTL                = 5 * time.Minute
	inventoryValuationCacheTTL      = 15 * time.Minute
	revenueByCategoryCacheTTL       = 15 * time.Minute
	orderStatusDistributionCacheTTL = 5 * time.Minute
)

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
	cached, err := a.getCachedTenantAnalytics(ctx, tenantID)
	if err != nil {
		log.Printf("Failed to read cached analytics for tenant %s: %v", tenantID.String(), err)
	} else if cached != nil {
		return cached, nil
	}

	data := &AnalyticsData{
		TenantID: tenantID,
	}

	invoices, err := a.invoiceRepo.List(ctx, tenantID, 10000, 0)
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

	inventories, err := a.inventoryRepo.List(ctx, tenantID, 10000, 0)
	if err != nil {
		log.Printf("Failed to get inventories for analytics: %v", err)
		return data, err
	}

	var totalStockValue float64
	lowStockCount := 0
	for _, inv := range inventories {
		if inv.Quantity < 10 {
			lowStockCount++
		}

		product, err := a.productRepo.GetByID(ctx, tenantID, inv.ProductID)
		if err != nil {
			log.Printf("Failed to get product %s: %v", inv.ProductID.String(), err)
			continue
		}
		value, err := common.SafeMultiplyMonetary(float64(inv.Quantity), product.UnitPrice)
		if err != nil {
			log.Printf("WARN: overflow computing stock value for product %s: %v", inv.ProductID.String(), err)
			continue
		}
		totalStockValue += value
	}

	data.TotalStockValue = totalStockValue
	data.LowStockItemsCount = lowStockCount
	data.LastUpdated = time.Now()

	a.cacheTenantAnalytics(ctx, data)

	return data, nil
}

func (a *AnalyticsService) GetSalesTrends(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]SalesTrend, error) {
	cacheKey := a.salesTrendsCacheKey(tenantID, startDate, endDate)
	var cached []SalesTrend
	if found, err := a.getCachedJSON(ctx, cacheKey, &cached); err != nil {
		log.Printf("Failed to read cached sales trends for tenant %s: %v", tenantID.String(), err)
	} else if found {
		return cached, nil
	}

	orders, err := a.orderRepo.GetOrdersByTenantAndDateRange(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	trends := make(map[string]*SalesTrend)
	for _, order := range orders {
		dateStr := order.OrderDate.Format("2006-01-02")
		if trends[dateStr] == nil {
			trends[dateStr] = &SalesTrend{Date: order.OrderDate}
		}
		amount, err := common.SafeMultiplyMonetary(float64(order.Quantity), order.UnitPrice)
		if err != nil {
			log.Printf("WARN: overflow computing sales trend for order %s: %v", order.ID, err)
		} else {
			trends[dateStr].SalesAmount += amount
		}
		trends[dateStr].OrderCount++
	}

	var result []SalesTrend
	for _, trend := range trends {
		result = append(result, *trend)
	}

	a.setCachedJSON(ctx, cacheKey, result, salesTrendCacheTTL)

	return result, nil
}

func (a *AnalyticsService) CalculateGSTTotals(ctx context.Context, tenantID uuid.UUID) (map[string]float64, error) {
	cacheKey := a.gstTotalsCacheKey(tenantID)
	var cached map[string]float64
	if found, err := a.getCachedJSON(ctx, cacheKey, &cached); err != nil {
		log.Printf("Failed to read cached GST totals for tenant %s: %v", tenantID.String(), err)
	} else if found {
		return cached, nil
	}

	invoices, err := a.invoiceRepo.List(ctx, tenantID, 10000, 0)
	if err != nil {
		return nil, err
	}

	totals := map[string]float64{
		"cgst":  0,
		"sgst":  0,
		"igst":  0,
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

	a.setCachedJSON(ctx, cacheKey, totals, gstTotalsCacheTTL)

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

	// Get most used filters from actual usage tracking
	// For now, return a computed list based on common patterns
	// TODO: Implement proper filter usage tracking table
	mostUsedFilters := []string{}
	
	// Query to check which filters are commonly used (basic heuristic)
	filterQuery := `
		SELECT 
			CASE 
				WHEN COUNT(DISTINCT category_id) > 10 THEN 'category'
				ELSE NULL
			END as filter_type
		FROM products 
		WHERE tenant_id = $1 AND deleted_at IS NULL
		UNION ALL
		SELECT 
			CASE 
				WHEN COUNT(*) > 50 THEN 'price_range'
				ELSE NULL
			END
		FROM products 
		WHERE tenant_id = $1 AND deleted_at IS NULL AND unit_price > 0
		UNION ALL
		SELECT 
			CASE 
				WHEN COUNT(DISTINCT warehouse_id) > 5 THEN 'warehouse'
				ELSE NULL
			END
		FROM inventory 
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`
	
	filterRows, err := a.db.Query(ctx, filterQuery, tenantID)
	if err == nil {
		defer filterRows.Close()
		for filterRows.Next() {
			var filterType *string
			if err := filterRows.Scan(&filterType); err == nil && filterType != nil {
				mostUsedFilters = append(mostUsedFilters, *filterType)
			}
		}
	}
	
	// Add default filters if none found
	if len(mostUsedFilters) == 0 {
		mostUsedFilters = []string{"category", "quantity", "price_range"}
	}
	
	metrics["most_used_filters"] = mostUsedFilters

	return metrics, nil
}

// InvalidateTenantAnalyticsCache invalidates cached analytics data for a tenant
func (a *AnalyticsService) InvalidateTenantAnalyticsCache(ctx context.Context, tenantID uuid.UUID) error {
	log.Printf("Invalidating analytics cache for tenant %s", tenantID.String())

	keys := []string{
		fmt.Sprintf("agromart:analytics:%s:dashboard", tenantID.String()),
		a.gstTotalsCacheKey(tenantID),
		a.inventoryValuationCacheKey(tenantID),
		a.revenueByCategoryCacheKey(tenantID),
		a.orderStatusDistributionCacheKey(tenantID),
	}

	var firstErr error
	for _, key := range keys {
		if err := a.cacheService.Delete(ctx, key); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// ProductSales represents product sales statistics
type ProductSales struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	TotalSales  float64   `json:"total_sales"`
	UnitsSold   int       `json:"units_sold"`
	OrderCount  int       `json:"order_count"`
}

// GetTopSellingProducts returns the top selling products by revenue
func (a *AnalyticsService) GetTopSellingProducts(ctx context.Context, tenantID uuid.UUID, limit int) ([]ProductSales, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Cap at 100
	}

	cacheKey := a.topProductsCacheKey(tenantID, limit)
	var cached []ProductSales
	if found, err := a.getCachedJSON(ctx, cacheKey, &cached); err != nil {
		log.Printf("Failed to read cached top products for tenant %s: %v", tenantID.String(), err)
	} else if found {
		return cached, nil
	}

	orders, err := a.orderRepo.List(ctx, tenantID, 100000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	productSalesMap := make(map[uuid.UUID]*ProductSales)

	for _, order := range orders {
		if order.Status == "cancelled" {
			continue
		}

		if _, exists := productSalesMap[order.ProductID]; !exists {
			product, err := a.productRepo.GetByID(ctx, tenantID, order.ProductID)
			if err != nil {
				log.Printf("Failed to get product %s: %v", order.ProductID.String(), err)
				continue
			}

			productSalesMap[order.ProductID] = &ProductSales{
				ProductID:   order.ProductID,
				ProductName: product.Name,
			}
		}

		sales := productSalesMap[order.ProductID]
		amount, err := common.SafeMultiplyMonetary(float64(order.Quantity), order.UnitPrice)
		if err != nil {
			log.Printf("WARN: overflow computing sales for product %s: %v", order.ProductID.String(), err)
			continue
		}
		sales.TotalSales += amount
		sales.UnitsSold += order.Quantity
		sales.OrderCount++
	}

	var productSales []ProductSales
	for _, ps := range productSalesMap {
		productSales = append(productSales, *ps)
	}

	for i := 0; i < len(productSales); i++ {
		for j := i + 1; j < len(productSales); j++ {
			if productSales[j].TotalSales > productSales[i].TotalSales {
				productSales[i], productSales[j] = productSales[j], productSales[i]
			}
		}
	}

	if len(productSales) > limit {
		productSales = productSales[:limit]
	}

	a.setCachedJSON(ctx, cacheKey, productSales, productSalesCacheTTL)

	log.Printf("Retrieved top %d selling products for tenant %s", len(productSales), tenantID.String())
	return productSales, nil
}

// LowStockItem represents a product with low inventory
type LowStockItem struct {
	ProductID    uuid.UUID `json:"product_id"`
	ProductName  string    `json:"product_name"`
	WarehouseID  uuid.UUID `json:"warehouse_id"`
	CurrentStock int       `json:"current_stock"`
	Threshold    int       `json:"threshold"`
	UnitPrice    float64   `json:"unit_price"`
	StockValue   float64   `json:"stock_value"`
}

// GetLowStockReport generates a report of products below the stock threshold
func (a *AnalyticsService) GetLowStockReport(ctx context.Context, tenantID uuid.UUID, threshold int) ([]LowStockItem, error) {
	if threshold <= 0 {
		threshold = 10 // Default threshold
	}

	cacheKey := a.lowStockCacheKey(tenantID, threshold)
	var cached []LowStockItem
	if found, err := a.getCachedJSON(ctx, cacheKey, &cached); err != nil {
		log.Printf("Failed to read cached low stock report for tenant %s: %v", tenantID.String(), err)
	} else if found {
		return cached, nil
	}

	inventories, err := a.inventoryRepo.List(ctx, tenantID, 100000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inventories: %w", err)
	}

	var lowStockItems []LowStockItem

	for _, inv := range inventories {
		if inv.Quantity >= threshold {
			continue
		}

		product, err := a.productRepo.GetByID(ctx, tenantID, inv.ProductID)
		if err != nil {
			log.Printf("Failed to get product %s: %v", inv.ProductID.String(), err)
			continue
		}

		stockValue, err := common.SafeMultiplyMonetary(float64(inv.Quantity), product.UnitPrice)
		if err != nil {
			log.Printf("WARN: overflow computing stock value for low stock product %s: %v", inv.ProductID.String(), err)
			continue
		}

		lowStockItems = append(lowStockItems, LowStockItem{
			ProductID:    inv.ProductID,
			ProductName:  product.Name,
			WarehouseID:  inv.WarehouseID,
			CurrentStock: inv.Quantity,
			Threshold:    threshold,
			UnitPrice:    product.UnitPrice,
			StockValue:   stockValue,
		})
	}

	for i := 0; i < len(lowStockItems); i++ {
		for j := i + 1; j < len(lowStockItems); j++ {
			if lowStockItems[j].CurrentStock < lowStockItems[i].CurrentStock {
				lowStockItems[i], lowStockItems[j] = lowStockItems[j], lowStockItems[i]
			}
		}
	}

	a.setCachedJSON(ctx, cacheKey, lowStockItems, lowStockCacheTTL)

	log.Printf("Generated low stock report for tenant %s: %d items below threshold %d",
		tenantID.String(), len(lowStockItems), threshold)

	return lowStockItems, nil
}

// InventoryValuation represents the total valuation of inventory
type InventoryValuation struct {
	TenantID       uuid.UUID          `json:"tenant_id"`
	TotalValue     float64            `json:"total_value"`
	TotalItems     int                `json:"total_items"`
	TotalQuantity  int                `json:"total_quantity"`
	ByWarehouse    map[string]float64 `json:"by_warehouse"`
	ByCategory     map[string]float64 `json:"by_category"`
	LastCalculated time.Time          `json:"last_calculated"`
}

// CalculateInventoryValuation calculates the total value of inventory
func (a *AnalyticsService) CalculateInventoryValuation(ctx context.Context, tenantID uuid.UUID) (*InventoryValuation, error) {
	cacheKey := a.inventoryValuationCacheKey(tenantID)
	var cached InventoryValuation
	if found, err := a.getCachedJSON(ctx, cacheKey, &cached); err != nil {
		log.Printf("Failed to read cached inventory valuation for tenant %s: %v", tenantID.String(), err)
	} else if found {
		return &cached, nil
	}

	valuation := &InventoryValuation{
		TenantID:       tenantID,
		TotalValue:     0,
		TotalItems:     0,
		TotalQuantity:  0,
		ByWarehouse:    make(map[string]float64),
		ByCategory:     make(map[string]float64),
		LastCalculated: time.Now(),
	}

	// Get all inventory items
	inventories, err := a.inventoryRepo.List(ctx, tenantID, 100000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inventories: %w", err)
	}

	for _, inv := range inventories {
		// Fetch product details
		product, err := a.productRepo.GetByID(ctx, tenantID, inv.ProductID)
		if err != nil {
			log.Printf("Failed to get product %s: %v", inv.ProductID.String(), err)
			continue
		}

		itemValue, err := common.SafeMultiplyMonetary(float64(inv.Quantity), product.UnitPrice)
		if err != nil {
			log.Printf("WARN: overflow computing valuation for product %s: %v", product.ID.String(), err)
			continue
		}
		valuation.TotalValue += itemValue
		valuation.TotalItems++
		valuation.TotalQuantity += inv.Quantity

		// Group by warehouse
		warehouseKey := inv.WarehouseID.String()
		valuation.ByWarehouse[warehouseKey] += itemValue

		// Group by category if available
		if product.CategoryID != nil {
			categoryKey := product.CategoryID.String()
			valuation.ByCategory[categoryKey] += itemValue
		}
	}

	a.setCachedJSON(ctx, cacheKey, valuation, inventoryValuationCacheTTL)

	log.Printf("Calculated inventory valuation for tenant %s: Total value: %.2f, Items: %d",
		tenantID.String(), valuation.TotalValue, valuation.TotalItems)

	return valuation, nil
}

// GetRevenueByCategory calculates revenue breakdown by product category
func (a *AnalyticsService) GetRevenueByCategory(ctx context.Context, tenantID uuid.UUID) (map[string]float64, error) {
	cacheKey := a.revenueByCategoryCacheKey(tenantID)
	var cached map[string]float64
	if found, err := a.getCachedJSON(ctx, cacheKey, &cached); err != nil {
		log.Printf("Failed to read cached revenue by category for tenant %s: %v", tenantID.String(), err)
	} else if found {
		return cached, nil
	}

	revenueByCategory := make(map[string]float64)

	orders, err := a.orderRepo.List(ctx, tenantID, 100000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	for _, order := range orders {
		if order.Status == "cancelled" {
			continue
		}

		product, err := a.productRepo.GetByID(ctx, tenantID, order.ProductID)
		if err != nil {
			log.Printf("Failed to get product %s: %v", order.ProductID.String(), err)
			continue
		}

		revenue, err := common.SafeMultiplyMonetary(float64(order.Quantity), order.UnitPrice)
		if err != nil {
			log.Printf("WARN: overflow computing revenue for order %s: %v", order.ID, err)
			continue
		}

		if product.CategoryID != nil {
			categoryKey := product.CategoryID.String()
			revenueByCategory[categoryKey] += revenue
		} else {
			revenueByCategory["uncategorized"] += revenue
		}
	}

	a.setCachedJSON(ctx, cacheKey, revenueByCategory, revenueByCategoryCacheTTL)

	log.Printf("Calculated revenue by category for tenant %s: %d categories",
		tenantID.String(), len(revenueByCategory))

	return revenueByCategory, nil
}

// GetOrderStatusDistribution returns the count of orders by status
func (a *AnalyticsService) GetOrderStatusDistribution(ctx context.Context, tenantID uuid.UUID) (map[string]int, error) {
	cacheKey := a.orderStatusDistributionCacheKey(tenantID)
	var cached map[string]int
	if found, err := a.getCachedJSON(ctx, cacheKey, &cached); err != nil {
		log.Printf("Failed to read cached order status distribution for tenant %s: %v", tenantID.String(), err)
	} else if found {
		return cached, nil
	}

	distribution := map[string]int{
		"pending":    0,
		"approved":   0,
		"processing": 0,
		"shipped":    0,
		"delivered":  0,
		"cancelled":  0,
	}

	orders, err := a.orderRepo.List(ctx, tenantID, 100000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	for _, order := range orders {
		distribution[order.Status]++
	}

	a.setCachedJSON(ctx, cacheKey, distribution, orderStatusDistributionCacheTTL)

	log.Printf("Calculated order status distribution for tenant %s: %d total orders",
		tenantID.String(), len(orders))

	return distribution, nil
}

// RefreshTenantAnalytics triggers a full refresh of analytics data for a tenant
func (a *AnalyticsService) RefreshTenantAnalytics(ctx context.Context, tenantID uuid.UUID) error {
	log.Printf("Starting analytics refresh for tenant %s", tenantID.String())

	if err := a.InvalidateTenantAnalyticsCache(ctx, tenantID); err != nil {
		log.Printf("Failed to invalidate analytics cache for tenant %s: %v", tenantID.String(), err)
	}

	if _, err := a.CalculateTenantAnalytics(ctx, tenantID); err != nil {
		return fmt.Errorf("failed to calculate tenant analytics: %w", err)
	}

	log.Printf("Analytics refresh completed for tenant %s", tenantID.String())
	return nil
}

func (a *AnalyticsService) getCachedTenantAnalytics(ctx context.Context, tenantID uuid.UUID) (*AnalyticsData, error) {
	cached, err := a.cacheService.GetTenantAnalytics(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if cached == nil {
		return nil, nil
	}

	data, err := analyticsDataFromCacheMap(cached)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (a *AnalyticsService) cacheTenantAnalytics(ctx context.Context, data *AnalyticsData) {
	if data == nil {
		return
	}

	payload := map[string]interface{}{
		"tenant_id":             data.TenantID.String(),
		"total_sales":           data.TotalSales,
		"total_stock_value":     data.TotalStockValue,
		"gst_collected":         data.GSTCollected,
		"order_count":           data.OrderCount,
		"low_stock_items_count": data.LowStockItemsCount,
		"last_updated":          data.LastUpdated.Format(time.RFC3339),
	}

	if err := a.cacheService.SetTenantAnalytics(ctx, data.TenantID, payload, defaultAnalyticsCacheTTL); err != nil {
		log.Printf("Failed to cache analytics for tenant %s: %v", data.TenantID.String(), err)
	}
}

func analyticsDataFromCacheMap(payload map[string]interface{}) (*AnalyticsData, error) {
	tenantStr, ok := payload["tenant_id"].(string)
	if !ok || tenantStr == "" {
		return nil, fmt.Errorf("invalid tenant_id in analytics cache")
	}

	tenantID, err := uuid.Parse(tenantStr)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id in analytics cache: %w", err)
	}

	data := &AnalyticsData{
		TenantID:           tenantID,
		TotalSales:         floatFromInterface(payload["total_sales"]),
		TotalStockValue:    floatFromInterface(payload["total_stock_value"]),
		GSTCollected:       floatFromInterface(payload["gst_collected"]),
		OrderCount:         intFromInterface(payload["order_count"]),
		LowStockItemsCount: intFromInterface(payload["low_stock_items_count"]),
		LastUpdated:        time.Now(),
	}

	if ts, ok := payload["last_updated"].(string); ok && ts != "" {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			data.LastUpdated = parsed
		}
	}

	return data, nil
}

func floatFromInterface(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f
		}
	case string:
		if v == "" {
			return 0
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}

	return 0
}

func intFromInterface(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i)
		}
		if f, err := v.Float64(); err == nil {
			return int(f)
		}
	case string:
		if v == "" {
			return 0
		}
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return int(f)
		}
	default:
		return int(floatFromInterface(v))
	}

	return 0
}

func (a *AnalyticsService) getCachedJSON(ctx context.Context, key string, dest interface{}) (bool, error) {
	cached, err := a.cacheService.GetString(ctx, key)
	if err != nil {
		return false, err
	}
	if cached == "" {
		return false, nil
	}

	if err := json.Unmarshal([]byte(cached), dest); err != nil {
		return false, err
	}

	return true, nil
}

func (a *AnalyticsService) setCachedJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("Failed to marshal analytics cache payload for key %s: %v", key, err)
		return
	}

	if err := a.cacheService.SetString(ctx, key, string(data), ttl); err != nil {
		log.Printf("Failed to cache analytics payload for key %s: %v", key, err)
	}
}

func (a *AnalyticsService) salesTrendsCacheKey(tenantID uuid.UUID, startDate, endDate time.Time) string {
	return fmt.Sprintf("agromart:analytics:%s:sales_trends:%s:%s", tenantID.String(), startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
}

func (a *AnalyticsService) gstTotalsCacheKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("agromart:analytics:%s:gst_totals", tenantID.String())
}

func (a *AnalyticsService) topProductsCacheKey(tenantID uuid.UUID, limit int) string {
	return fmt.Sprintf("agromart:analytics:%s:top_products:%d", tenantID.String(), limit)
}

func (a *AnalyticsService) lowStockCacheKey(tenantID uuid.UUID, threshold int) string {
	return fmt.Sprintf("agromart:analytics:%s:low_stock:%d", tenantID.String(), threshold)
}

func (a *AnalyticsService) inventoryValuationCacheKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("agromart:analytics:%s:inventory_valuation", tenantID.String())
}

func (a *AnalyticsService) revenueByCategoryCacheKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("agromart:analytics:%s:revenue_by_category", tenantID.String())
}

func (a *AnalyticsService) orderStatusDistributionCacheKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("agromart:analytics:%s:order_status_distribution", tenantID.String())
}
