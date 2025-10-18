package handlers

import (
	"net/http"
	"strconv"

	"agromart2/internal/common"
	"agromart2/internal/middleware"

	"github.com/labstack/echo/v4"
)

// DBOptimizationHandlers handles database optimization endpoints
type DBOptimizationHandlers struct {
	optimizer      *common.DBOptimizer
	rbacMiddleware *middleware.RBACMiddleware
}

// NewDBOptimizationHandlers creates a new database optimization handlers instance
func NewDBOptimizationHandlers(optimizer *common.DBOptimizer, rbacMiddleware *middleware.RBACMiddleware) *DBOptimizationHandlers {
	return &DBOptimizationHandlers{
		optimizer:      optimizer,
		rbacMiddleware: rbacMiddleware,
	}
}

// RefreshMaterializedViews refreshes all materialized views
func (h *DBOptimizationHandlers) RefreshMaterializedViews(c echo.Context) error {
	ctx := c.Request().Context()

	if err := h.optimizer.RefreshMaterializedViews(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to refresh materialized views: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Materialized views refreshed successfully",
	})
}

// GetSlowQueries returns slow queries from pg_stat_statements
func (h *DBOptimizationHandlers) GetSlowQueries(c echo.Context) error {
	ctx := c.Request().Context()

	// Get limit from query params (default 20)
	limitStr := c.QueryParam("limit")
	limit := 20
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	queries, err := h.optimizer.GetSlowQueries(ctx, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get slow queries: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"slow_queries": queries,
		"count":        len(queries),
	})
}

// GetUnusedIndexes returns indexes that are never used
func (h *DBOptimizationHandlers) GetUnusedIndexes(c echo.Context) error {
	ctx := c.Request().Context()

	indexes, err := h.optimizer.GetUnusedIndexes(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get unused indexes: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"unused_indexes": indexes,
		"count":          len(indexes),
	})
}

// GetTableSizes returns sizes of all tables
func (h *DBOptimizationHandlers) GetTableSizes(c echo.Context) error {
	ctx := c.Request().Context()

	sizes, err := h.optimizer.GetTableSizes(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get table sizes: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"table_sizes": sizes,
		"count":       len(sizes),
	})
}

// GetCacheHitRatio returns the cache hit ratio
func (h *DBOptimizationHandlers) GetCacheHitRatio(c echo.Context) error {
	ctx := c.Request().Context()

	ratio, err := h.optimizer.GetCacheHitRatio(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get cache hit ratio: "+err.Error())
	}

	// Determine health status
	status := "excellent"
	if ratio < 0.99 {
		status = "good"
	}
	if ratio < 0.95 {
		status = "warning"
	}
	if ratio < 0.90 {
		status = "critical"
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"cache_hit_ratio": ratio,
		"percentage":      ratio * 100,
		"status":          status,
		"recommendation":  getRecommendation(ratio),
	})
}

// VacuumAnalyze performs VACUUM ANALYZE on specified tables
func (h *DBOptimizationHandlers) VacuumAnalyze(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		Tables []string `json:"tables" validate:"required,min=1"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.optimizer.VacuumAnalyze(ctx, req.Tables); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to vacuum analyze: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "VACUUM ANALYZE completed successfully",
		"tables":  req.Tables,
	})
}

// GetDatabaseStats returns comprehensive database statistics
func (h *DBOptimizationHandlers) GetDatabaseStats(c echo.Context) error {
	ctx := c.Request().Context()

	// Get cache hit ratio
	cacheRatio, err := h.optimizer.GetCacheHitRatio(ctx)
	if err != nil {
		cacheRatio = 0
	}

	// Get table sizes
	tableSizes, err := h.optimizer.GetTableSizes(ctx)
	if err != nil {
		tableSizes = []common.TableSize{}
	}

	// Get slow queries
	slowQueries, err := h.optimizer.GetSlowQueries(ctx, 10)
	if err != nil {
		slowQueries = []common.SlowQuery{}
	}

	// Get unused indexes
	unusedIndexes, err := h.optimizer.GetUnusedIndexes(ctx)
	if err != nil {
		unusedIndexes = []common.UnusedIndex{}
	}

	// Calculate total database size
	var totalSize int64
	for _, table := range tableSizes {
		totalSize += table.SizeBytes
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"cache_hit_ratio":  cacheRatio,
		"cache_percentage": cacheRatio * 100,
		"cache_status":     getCacheStatus(cacheRatio),
		"total_tables":     len(tableSizes),
		"total_size_bytes": totalSize,
		"slow_queries":     len(slowQueries),
		"unused_indexes":   len(unusedIndexes),
		"top_tables":       getTopTables(tableSizes, 10),
		"recommendations":  generateRecommendations(cacheRatio, slowQueries, unusedIndexes),
	})
}

// Helper functions

func getRecommendation(ratio float64) string {
	if ratio >= 0.99 {
		return "Excellent cache performance. No action needed."
	} else if ratio >= 0.95 {
		return "Good cache performance. Consider increasing shared_buffers if memory allows."
	} else if ratio >= 0.90 {
		return "Cache performance needs improvement. Increase shared_buffers and effective_cache_size."
	}
	return "Critical: Poor cache performance. Immediately increase shared_buffers and review query patterns."
}

func getCacheStatus(ratio float64) string {
	if ratio >= 0.99 {
		return "excellent"
	} else if ratio >= 0.95 {
		return "good"
	} else if ratio >= 0.90 {
		return "warning"
	}
	return "critical"
}

func getTopTables(sizes []common.TableSize, limit int) []common.TableSize {
	if len(sizes) <= limit {
		return sizes
	}
	return sizes[:limit]
}

func generateRecommendations(cacheRatio float64, slowQueries []common.SlowQuery, unusedIndexes []common.UnusedIndex) []string {
	recommendations := []string{}

	// Cache recommendations
	if cacheRatio < 0.99 {
		recommendations = append(recommendations, "Increase shared_buffers to improve cache hit ratio")
	}

	// Slow query recommendations
	if len(slowQueries) > 0 {
		recommendations = append(recommendations, "Review and optimize slow queries")
		if len(slowQueries) > 10 {
			recommendations = append(recommendations, "Consider adding indexes for frequently slow queries")
		}
	}

	// Unused index recommendations
	if len(unusedIndexes) > 0 {
		recommendations = append(recommendations, "Consider dropping unused indexes to improve write performance")
	}

	// General recommendations
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Database is well optimized. Continue monitoring.")
	}

	return recommendations
}
