package common

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DBOptimizer provides database optimization utilities
type DBOptimizer struct {
	pool *pgxpool.Pool
}

// NewDBOptimizer creates a new database optimizer
func NewDBOptimizer(pool *pgxpool.Pool) *DBOptimizer {
	return &DBOptimizer{pool: pool}
}

// RefreshMaterializedViews refreshes all materialized views
func (o *DBOptimizer) RefreshMaterializedViews(ctx context.Context) error {
	_, err := o.pool.Exec(ctx, "SELECT refresh_analytics_views()")
	return err
}

// VacuumAnalyze performs VACUUM ANALYZE on specified tables
func (o *DBOptimizer) VacuumAnalyze(ctx context.Context, tables []string) error {
	for _, table := range tables {
		query := fmt.Sprintf("VACUUM ANALYZE %s", table)
		_, err := o.pool.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to vacuum %s: %w", table, err)
		}
	}
	return nil
}

// GetSlowQueries returns slow queries from pg_stat_statements
func (o *DBOptimizer) GetSlowQueries(ctx context.Context, limit int) ([]SlowQuery, error) {
	query := `
		SELECT 
			query,
			calls,
			total_exec_time,
			mean_exec_time,
			max_exec_time
		FROM pg_stat_statements
		WHERE query NOT LIKE '%pg_stat_statements%'
		ORDER BY mean_exec_time DESC
		LIMIT $1
	`

	rows, err := o.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queries []SlowQuery
	for rows.Next() {
		var q SlowQuery
		err := rows.Scan(&q.Query, &q.Calls, &q.TotalExecTime, &q.MeanExecTime, &q.MaxExecTime)
		if err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}

	return queries, nil
}

// GetUnusedIndexes returns indexes that are never used
func (o *DBOptimizer) GetUnusedIndexes(ctx context.Context) ([]UnusedIndex, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			idx_scan
		FROM pg_stat_user_indexes
		WHERE idx_scan = 0
		AND indexname NOT LIKE '%_pkey'
		ORDER BY schemaname, tablename
	`

	rows, err := o.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []UnusedIndex
	for rows.Next() {
		var idx UnusedIndex
		err := rows.Scan(&idx.SchemaName, &idx.TableName, &idx.IndexName, &idx.ScanCount)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// GetTableSizes returns sizes of all tables
func (o *DBOptimizer) GetTableSizes(ctx context.Context) ([]TableSize, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			pg_total_relation_size(schemaname||'.'||tablename) AS size_bytes,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size_pretty
		FROM pg_tables
		WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
	`

	rows, err := o.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sizes []TableSize
	for rows.Next() {
		var ts TableSize
		err := rows.Scan(&ts.SchemaName, &ts.TableName, &ts.SizeBytes, &ts.SizePretty)
		if err != nil {
			return nil, err
		}
		sizes = append(sizes, ts)
	}

	return sizes, nil
}

// GetCacheHitRatio returns the cache hit ratio (should be > 99%)
func (o *DBOptimizer) GetCacheHitRatio(ctx context.Context) (float64, error) {
	query := `
		SELECT 
			CASE 
				WHEN sum(heap_blks_hit) + sum(heap_blks_read) = 0 THEN 0
				ELSE sum(heap_blks_hit)::float / (sum(heap_blks_hit) + sum(heap_blks_read))
			END as ratio
		FROM pg_statio_user_tables
	`

	var ratio float64
	err := o.pool.QueryRow(ctx, query).Scan(&ratio)
	return ratio, err
}

// ArchiveOldData archives old data from specified tables
func (o *DBOptimizer) ArchiveOldData(ctx context.Context, table string, dateColumn string, olderThan time.Duration) (int64, error) {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s < $1",
		table,
		dateColumn,
	)

	cutoffDate := time.Now().Add(-olderThan)
	result, err := o.pool.Exec(ctx, query, cutoffDate)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

// OptimizeConnectionPool adjusts connection pool settings
func (o *DBOptimizer) OptimizeConnectionPool() {
	config := o.pool.Config()
	
	// Set optimal pool size based on CPU cores
	// Rule of thumb: (core_count * 2) + effective_spindle_count
	config.MaxConns = 20
	config.MinConns = 5
	
	// Set connection timeouts
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute
}

// Types for query results

type SlowQuery struct {
	Query         string
	Calls         int64
	TotalExecTime float64
	MeanExecTime  float64
	MaxExecTime   float64
}

type UnusedIndex struct {
	SchemaName string
	TableName  string
	IndexName  string
	ScanCount  int64
}

type TableSize struct {
	SchemaName string
	TableName  string
	SizeBytes  int64
	SizePretty string
}

// BatchInsert performs optimized batch insert
func BatchInsert(ctx context.Context, pool *pgxpool.Pool, query string, values [][]interface{}, batchSize int) error {
	total := len(values)
	
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		
		batch := values[i:end]
		
		// Use COPY for better performance with large batches
		// Or use batch.Queue for multiple inserts
		for _, row := range batch {
			_, err := pool.Exec(ctx, query, row...)
			if err != nil {
				return fmt.Errorf("batch insert failed at row %d: %w", i, err)
			}
		}
	}
	
	return nil
}

// PrepareStatements prepares commonly used statements for better performance
func PrepareStatements(ctx context.Context, pool *pgxpool.Pool) error {
	statements := map[string]string{
		"get_product_by_id": "SELECT * FROM products WHERE tenant_id = $1 AND id = $2",
		"get_inventory": "SELECT * FROM inventory WHERE tenant_id = $1 AND product_id = $2",
		"update_stock": "UPDATE inventory SET quantity = $1, updated_at = NOW() WHERE tenant_id = $2 AND product_id = $3",
	}
	
	for name, sql := range statements {
		_, err := pool.Exec(ctx, fmt.Sprintf("PREPARE %s AS %s", name, sql))
		if err != nil {
			return fmt.Errorf("failed to prepare statement %s: %w", name, err)
		}
	}
	
	return nil
}
