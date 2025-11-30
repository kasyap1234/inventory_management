package handlers

import (
	"agromart2/internal/analytics"
	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// AnalyticsHandlers handles HTTP requests for analytics and reporting
type AnalyticsHandlers struct {
	analyticsSvc   *analytics.AnalyticsService
	rbacMiddleware *middleware.RBACMiddleware
}

// NewAnalyticsHandlers creates a new analytics handlers instance
func NewAnalyticsHandlers(analyticsSvc *analytics.AnalyticsService, rbacMiddleware *middleware.RBACMiddleware) *AnalyticsHandlers {
	return &AnalyticsHandlers{
		analyticsSvc:   analyticsSvc,
		rbacMiddleware: rbacMiddleware,
	}
}

// GetDashboardAnalytics handles GET /analytics/dashboard
// Returns key metrics for the tenant dashboard
func (h *AnalyticsHandlers) GetDashboardAnalytics(c echo.Context) error {
	ctx := c.Request().Context()

	// Check RBAC permission
	err := h.rbacMiddleware.RequirePermission("analytics.read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to view analytics")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Calculate analytics
	data, err := h.analyticsSvc.CalculateTenantAnalytics(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to calculate analytics: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tenant_id":         data.TenantID,
		"total_sales":       data.TotalSales,
		"total_stock_value": data.TotalStockValue,
		"gst_collected":     data.GSTCollected,
		"order_count":       data.OrderCount,
		"low_stock_items":   data.LowStockItemsCount,
		"last_updated":      data.LastUpdated,
	})
}

// GetSalesTrends handles GET /analytics/sales-trends
// Returns sales trends over a date range
func (h *AnalyticsHandlers) GetSalesTrends(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Parse query parameters for date range
	startDateStr := c.QueryParam("start_date")
	endDateStr := c.QueryParam("end_date")

	// Default to last 30 days if not provided
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid start_date format (expected YYYY-MM-DD)")
		}
		startDate = parsed
	}

	if endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid end_date format (expected YYYY-MM-DD)")
		}
		endDate = parsed
	}

	// Get trends
	trends, err := h.analyticsSvc.GetSalesTrends(ctx, tenantID, startDate, endDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get sales trends: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"trends":     trends,
		"count":      len(trends),
	})
}

// GetGSTTotals handles GET /analytics/gst-totals
// Returns GST collection totals
func (h *AnalyticsHandlers) GetGSTTotals(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	totals, err := h.analyticsSvc.CalculateGSTTotals(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to calculate GST totals: "+err.Error())
	}

	return c.JSON(http.StatusOK, totals)
}

// GetTopProducts handles GET /analytics/top-products
// Returns top selling products
func (h *AnalyticsHandlers) GetTopProducts(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Default limit to 10
	limit := 10
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l := common.ParseIntOrDefault(limitParam, 10); l > 0 && l <= 100 {
			limit = l
		}
	}

	products, err := h.analyticsSvc.GetTopSellingProducts(ctx, tenantID, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get top products: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"products": products,
		"limit":    limit,
	})
}

// GetLowStockReport handles GET /analytics/low-stock
// Returns products with low stock levels
func (h *AnalyticsHandlers) GetLowStockReport(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Default threshold to 10
	threshold := 10
	if thresholdParam := c.QueryParam("threshold"); thresholdParam != "" {
		if t := common.ParseIntOrDefault(thresholdParam, 10); t > 0 {
			threshold = t
		}
	}

	report, err := h.analyticsSvc.GetLowStockReport(ctx, tenantID, threshold)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get low stock report: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"low_stock_items": report,
		"threshold":       threshold,
		"count":           len(report),
	})
}

// GetInventoryValuation handles GET /analytics/inventory-valuation
// Returns total inventory valuation
func (h *AnalyticsHandlers) GetInventoryValuation(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	valuation, err := h.analyticsSvc.CalculateInventoryValuation(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to calculate inventory valuation: "+err.Error())
	}

	return c.JSON(http.StatusOK, valuation)
}

// GetRevenueByCategory handles GET /analytics/revenue-by-category
// Returns revenue breakdown by product category
func (h *AnalyticsHandlers) GetRevenueByCategory(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	revenue, err := h.analyticsSvc.GetRevenueByCategory(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get revenue by category: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"revenue_by_category": revenue,
	})
}

// GetOrderStatusDistribution handles GET /analytics/order-status
// Returns distribution of orders by status
func (h *AnalyticsHandlers) GetOrderStatusDistribution(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	distribution, err := h.analyticsSvc.GetOrderStatusDistribution(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get order status distribution: "+err.Error())
	}

	return c.JSON(http.StatusOK, distribution)
}

// GetCustomerSegmentation handles GET /analytics/customer-segmentation
// Returns customer segmentation analytics
// NOTE: This handler is currently commented out because the underlying service method is disabled
// The Order model doesn't have CustomerID (B2B system with suppliers/distributors)
/*
func (h *AnalyticsHandlers) GetCustomerSegmentation(c echo.Context) error {
	ctx := c.Request().Context()

	// Check RBAC permission
	err := h.rbacMiddleware.RequirePermission("analytics.read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to view analytics")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	segmentation, err := h.analyticsSvc.GetCustomerSegmentation(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get customer segmentation: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"segmentation": segmentation,
	})
}
*/

// GetProductPerformance handles GET /analytics/product-performance
// Returns product performance metrics
func (h *AnalyticsHandlers) GetProductPerformance(c echo.Context) error {
	ctx := c.Request().Context()

	// Check RBAC permission
	err := h.rbacMiddleware.RequirePermission("analytics.read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to view analytics")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	limit := 20
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l := common.ParseIntOrDefault(limitParam, 20); l > 0 && l <= 100 {
			limit = l
		}
	}

	performance, err := h.analyticsSvc.GetProductPerformance(ctx, tenantID, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get product performance: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"performance": performance,
		"limit":       limit,
	})
}

// GetInventoryTurnover handles GET /analytics/inventory-turnover
// Returns inventory turnover analytics
func (h *AnalyticsHandlers) GetInventoryTurnover(c echo.Context) error {
	ctx := c.Request().Context()

	// Check RBAC permission
	err := h.rbacMiddleware.RequirePermission("analytics.read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to view analytics")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	turnover, err := h.analyticsSvc.GetInventoryTurnover(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get inventory turnover: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"turnover": turnover,
	})
}

// GetSupplierPerformance handles GET /analytics/supplier-performance
// Returns supplier performance metrics
func (h *AnalyticsHandlers) GetSupplierPerformance(c echo.Context) error {
	ctx := c.Request().Context()

	// Check RBAC permission
	err := h.rbacMiddleware.RequirePermission("analytics.read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to view analytics")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	performance, err := h.analyticsSvc.GetSupplierPerformance(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get supplier performance: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"performance": performance,
	})
}

// GetCustomerLifetimeValue handles GET /analytics/customer-lifetime-value
// Returns customer lifetime value analytics
// NOTE: This handler is currently commented out because the underlying service method is disabled
// The Order model doesn't have CustomerID (B2B system with suppliers/distributors)
/*
func (h *AnalyticsHandlers) GetCustomerLifetimeValue(c echo.Context) error {
	ctx := c.Request().Context()

	// Check RBAC permission
	err := h.rbacMiddleware.RequirePermission("analytics.read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to view analytics")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	limit := 20
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l := common.ParseIntOrDefault(limitParam, 20); l > 0 && l <= 100 {
			limit = l
		}
	}

	ltv, err := h.analyticsSvc.GetCustomerLifetimeValue(ctx, tenantID, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get customer lifetime value: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"customer_ltv": ltv,
		"limit":        limit,
	})
}
*/

// GetMarketTrends handles GET /analytics/market-trends
// Returns market trend analytics
func (h *AnalyticsHandlers) GetMarketTrends(c echo.Context) error {
	ctx := c.Request().Context()

	// Check RBAC permission
	err := h.rbacMiddleware.RequirePermission("analytics.read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to view analytics")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	trends, err := h.analyticsSvc.GetMarketTrends(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get market trends: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"trends": trends,
	})
}

// GetProfitMarginAnalysis handles GET /analytics/profit-margin
// Returns profit margin analysis
func (h *AnalyticsHandlers) GetProfitMarginAnalysis(c echo.Context) error {
	ctx := c.Request().Context()

	// Check RBAC permission
	err := h.rbacMiddleware.RequirePermission("analytics.read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to view analytics")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	analysis, err := h.analyticsSvc.GetProfitMarginAnalysis(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get profit margin analysis: "+err.Error())
	}

	return c.JSON(http.StatusOK, analysis)
}

// RefreshAnalytics handles POST /analytics/refresh
// Manually triggers analytics refresh for the tenant
func (h *AnalyticsHandlers) RefreshAnalytics(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Trigger async refresh
	err := h.analyticsSvc.RefreshTenantAnalytics(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to refresh analytics: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Analytics refresh initiated successfully",
	})
}

// CombinedAnalyticsResponse holds all analytics data in a single response
type CombinedAnalyticsResponse struct {
	Dashboard           interface{} `json:"dashboard"`
	SalesTrends         interface{} `json:"sales_trends"`
	TopProducts         interface{} `json:"top_products"`
	LowStock            interface{} `json:"low_stock"`
	InventoryValuation  interface{} `json:"inventory_valuation"`
	RevenueByCategory   interface{} `json:"revenue_by_category"`
	OrderStatus         interface{} `json:"order_status"`
	GSTTotals           interface{} `json:"gst_totals"`
	FetchedAt           time.Time   `json:"fetched_at"`
}

// GetCombinedAnalytics handles GET /analytics/combined
// Returns all analytics data in a single API call to reduce frontend round trips
// This endpoint fetches dashboard, sales trends, top products, low stock, inventory valuation,
// revenue by category, order status distribution, and GST totals in parallel
func (h *AnalyticsHandlers) GetCombinedAnalytics(c echo.Context) error {
	ctx := c.Request().Context()

	// Check RBAC permission
	err := h.rbacMiddleware.RequirePermission("analytics.read")(func(c echo.Context) error {
		return nil
	})(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to view analytics")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Parse optional parameters
	startDateStr := c.QueryParam("start_date")
	endDateStr := c.QueryParam("end_date")
	
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30) // Default to last 30 days

	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	topProductsLimit := common.ParseIntOrDefault(c.QueryParam("top_products_limit"), 10)
	lowStockThreshold := common.ParseIntOrDefault(c.QueryParam("low_stock_threshold"), 10)

	// Use WaitGroup to fetch all data in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex

	response := &CombinedAnalyticsResponse{
		FetchedAt: time.Now(),
	}

	// Fetch dashboard analytics
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := h.analyticsSvc.CalculateTenantAnalytics(ctx, tenantID)
		if err == nil {
			mu.Lock()
			response.Dashboard = map[string]interface{}{
				"tenant_id":         data.TenantID,
				"total_sales":       data.TotalSales,
				"total_stock_value": data.TotalStockValue,
				"gst_collected":     data.GSTCollected,
				"order_count":       data.OrderCount,
				"low_stock_items":   data.LowStockItemsCount,
				"last_updated":      data.LastUpdated,
			}
			mu.Unlock()
		}
	}()

	// Fetch sales trends
	wg.Add(1)
	go func() {
		defer wg.Done()
		trends, err := h.analyticsSvc.GetSalesTrends(ctx, tenantID, startDate, endDate)
		if err == nil {
			mu.Lock()
			response.SalesTrends = map[string]interface{}{
				"start_date": startDate.Format("2006-01-02"),
				"end_date":   endDate.Format("2006-01-02"),
				"trends":     trends,
				"count":      len(trends),
			}
			mu.Unlock()
		}
	}()

	// Fetch top products
	wg.Add(1)
	go func() {
		defer wg.Done()
		products, err := h.analyticsSvc.GetTopSellingProducts(ctx, tenantID, topProductsLimit)
		if err == nil {
			mu.Lock()
			response.TopProducts = map[string]interface{}{
				"products": products,
				"limit":    topProductsLimit,
			}
			mu.Unlock()
		}
	}()

	// Fetch low stock report
	wg.Add(1)
	go func() {
		defer wg.Done()
		report, err := h.analyticsSvc.GetLowStockReport(ctx, tenantID, lowStockThreshold)
		if err == nil {
			mu.Lock()
			response.LowStock = map[string]interface{}{
				"low_stock_items": report,
				"threshold":       lowStockThreshold,
				"count":           len(report),
			}
			mu.Unlock()
		}
	}()

	// Fetch inventory valuation
	wg.Add(1)
	go func() {
		defer wg.Done()
		valuation, err := h.analyticsSvc.CalculateInventoryValuation(ctx, tenantID)
		if err == nil {
			mu.Lock()
			response.InventoryValuation = valuation
			mu.Unlock()
		}
	}()

	// Fetch revenue by category
	wg.Add(1)
	go func() {
		defer wg.Done()
		revenue, err := h.analyticsSvc.GetRevenueByCategory(ctx, tenantID)
		if err == nil {
			mu.Lock()
			response.RevenueByCategory = map[string]interface{}{
				"revenue_by_category": revenue,
			}
			mu.Unlock()
		}
	}()

	// Fetch order status distribution
	wg.Add(1)
	go func() {
		defer wg.Done()
		distribution, err := h.analyticsSvc.GetOrderStatusDistribution(ctx, tenantID)
		if err == nil {
			mu.Lock()
			response.OrderStatus = distribution
			mu.Unlock()
		}
	}()

	// Fetch GST totals
	wg.Add(1)
	go func() {
		defer wg.Done()
		totals, err := h.analyticsSvc.CalculateGSTTotals(ctx, tenantID)
		if err == nil {
			mu.Lock()
			response.GSTTotals = totals
			mu.Unlock()
		}
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	return c.JSON(http.StatusOK, response)
}
