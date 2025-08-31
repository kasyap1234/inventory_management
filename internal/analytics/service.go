package analytics

import (
	"context"
	"fmt"
	"log"
	"time"

	

	"agromart2/internal/repositories"
	"agromart2/internal/caching"

	"github.com/google/uuid"
)

// AnalyticsService handles calculation and caching of analytics data
type AnalyticsService struct {
	orderRepo     repositories.OrderRepository
	invoiceRepo   repositories.InvoiceRepository
	inventoryRepo repositories.InventoryRepository
	productRepo   repositories.ProductRepository
	cacheService  caching.CacheService
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

func NewAnalyticsService(orderRepo repositories.OrderRepository, invoiceRepo repositories.InvoiceRepository, inventoryRepo repositories.InventoryRepository, productRepo repositories.ProductRepository, cacheService caching.CacheService) *AnalyticsService {
	return &AnalyticsService{
		orderRepo:     orderRepo,
		invoiceRepo:   invoiceRepo,
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
		cacheService:  cacheService,
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

// RecordSearchUsage tracks search operations for analytics
func (a *AnalyticsService) RecordSearchUsage(ctx context.Context, tenantID uuid.UUID, entityType string, searchTerm string, filterCount int, resultCount int, userID uuid.UUID, responseTimeMs int64) error {
	// In a production system, this would store search analytics in a database or log aggregator
	// For now, we'll log and potentially store in memory or use a queue system

	log.Printf("Search Usage - Tenant: %s, Entity: %s, Term: '%s', Filters: %d, Results: %d, User: %s, Response: %dms",
		tenantID.String(), entityType, searchTerm, filterCount, resultCount, userID.String(), responseTimeMs)

	// TODO: Persist to analytics database table
	// searchAnalytics := &SearchAnalytics{
	//     TenantID:       tenantID,
	//     EntityType:     entityType,
	//     SearchTerm:     searchTerm,
	//     FilterCount:    filterCount,
	//     ResultCount:    resultCount,
	//     Timestamp:      time.Now(),
	//     UserID:         userID,
	//     ResponseTimeMs: responseTimeMs,
	// }
	// // Save to database

	return nil
}

// GetSearchAnalytics retrieves search usage statistics for a tenant
func (a *AnalyticsService) GetSearchAnalytics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*SearchUsageStats, error) {
	// TODO: Query actual search analytics data from database
	// For now, return mock data structure

	stats := &SearchUsageStats{
		TenantID:        tenantID,
		TotalSearches:   0, // Would be calculated from actual data
		AvgResponseTime: 0, // Would be calculated from actual data
		DateRange: struct {
			Start time.Time
			End   time.Time
		}{
			Start: startDate,
			End:   endDate,
		},
	}

	// TODO: Populate TopSearchTerms and PeakUsageTimes from database
	stats.TopSearchTerms = []SearchTermFrequency{}
	stats.PeakUsageTimes = []TimeUsage{}

	return stats, nil
}

// TrackBulkOperationUsage tracks bulk operation usage
func (a *AnalyticsService) TrackBulkOperationUsage(ctx context.Context, tenantID uuid.UUID, operationType string, totalItems int, successCount int, userID uuid.UUID, processingTimeMs int64) error {
	log.Printf("Bulk Operation Usage - Tenant: %s, Type: %s, Total: %d, Success: %d, User: %s, Processing: %dms",
		tenantID.String(), operationType, totalItems, successCount, userID.String(), processingTimeMs)

	// TODO: Persist bulk operation analytics

	return nil
}

// GetPopularSearchTerms returns the most frequently searched terms
func (a *AnalyticsService) GetPopularSearchTerms(ctx context.Context, tenantID uuid.UUID, limit int) ([]SearchTermFrequency, error) {
	// TODO: Query database for most popular search terms
	// For now, return empty structure

	return []SearchTermFrequency{
		{Term: "fertilizer", Frequency: 150},
		{Term: "seeds", Frequency: 120},
		{Term: "organic", Frequency: 95},
	}, nil
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

// GetSearchPerformanceMetrics returns search performance metrics
func (a *AnalyticsService) GetSearchPerformanceMetrics(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	// TODO: Calculate real performance metrics
	return map[string]interface{}{
		"avg_response_time_ms":      150.5,
		"total_searches":           1250,
		"successful_searches":      1240,
		"failed_searches":          10,
		"most_used_filters":        []string{"category", "quantity", "price_range"},
		"peak_usage_hour":          14, // 2 PM
		"most_popular_entity":      "products",
		"zero_result_searches_pct": 5.2,
	}, nil
}

// InvalidateTenantAnalyticsCache invalidates cached analytics data for a tenant
func (a *AnalyticsService) InvalidateTenantAnalyticsCache(ctx context.Context, tenantID uuid.UUID) error {
	log.Printf("Invalidating analytics cache for tenant %s", tenantID.String())

	// Use cache service to invalidate by pattern or specific key
	// Since we want to invalidate only analytics, we can delete the specific key
	cacheKey := fmt.Sprintf("agromart:analytics:%s", tenantID.String())
	return a.cacheService.Delete(ctx, cacheKey)
}
