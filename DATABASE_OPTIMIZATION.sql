-- ============================================
-- DATABASE PERFORMANCE OPTIMIZATION
-- ============================================
-- Run these queries to optimize database performance

-- 1. CREATE ESSENTIAL INDEXES
-- ============================================

-- Products table indexes
CREATE INDEX IF NOT EXISTS idx_products_tenant_id ON products(tenant_id);
CREATE INDEX IF NOT EXISTS idx_products_barcode ON products(barcode) WHERE barcode IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id) WHERE category_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_products_name_search ON products USING gin(to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(tenant_id, created_at DESC);

-- Inventory table indexes
CREATE INDEX IF NOT EXISTS idx_inventory_product_id ON inventory(product_id);
CREATE INDEX IF NOT EXISTS idx_inventory_warehouse_id ON inventory(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_product ON inventory(tenant_id, product_id);
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_warehouse ON inventory(tenant_id, warehouse_id);
CREATE INDEX IF NOT EXISTS idx_inventory_low_stock ON inventory(tenant_id, quantity) WHERE quantity <= minimum_level;

-- Orders table indexes
CREATE INDEX IF NOT EXISTS idx_orders_tenant_id ON orders(tenant_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_tenant_status ON orders(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id) WHERE customer_id IS NOT NULL;

-- Order items indexes
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);

-- Invoices table indexes
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_id ON invoices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_status ON invoices(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(tenant_id, due_date) WHERE status != 'paid';
CREATE INDEX IF NOT EXISTS idx_invoices_order_id ON invoices(order_id) WHERE order_id IS NOT NULL;

-- Inventory reservations indexes
CREATE INDEX IF NOT EXISTS idx_reservations_product_id ON inventory_reservations(product_id);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON inventory_reservations(status);
CREATE INDEX IF NOT EXISTS idx_reservations_tenant_status ON inventory_reservations(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_reservations_reservation_id ON inventory_reservations(reservation_id);

-- Stock adjustments indexes
CREATE INDEX IF NOT EXISTS idx_stock_adjustments_product_id ON stock_adjustments(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_adjustments_tenant_product ON stock_adjustments(tenant_id, product_id);
CREATE INDEX IF NOT EXISTS idx_stock_adjustments_adjusted_at ON stock_adjustments(adjusted_at DESC);

-- Users table indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);

-- Audit logs indexes
CREATE INDEX IF NOT EXISTS idx_audit_logs_table_name ON audit_logs(table_name);
CREATE INDEX IF NOT EXISTS idx_audit_logs_record_id ON audit_logs(record_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_table ON audit_logs(tenant_id, table_name) WHERE tenant_id IS NOT NULL;

-- Categories indexes
CREATE INDEX IF NOT EXISTS idx_categories_tenant_id ON categories(tenant_id);
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id) WHERE parent_id IS NOT NULL;

-- Warehouses indexes
CREATE INDEX IF NOT EXISTS idx_warehouses_tenant_id ON warehouses(tenant_id);

-- 2. ANALYZE TABLES FOR QUERY PLANNER
-- ============================================
ANALYZE products;
ANALYZE inventory;
ANALYZE orders;
ANALYZE order_items;
ANALYZE invoices;
ANALYZE inventory_reservations;
ANALYZE stock_adjustments;
ANALYZE users;
ANALYZE audit_logs;
ANALYZE categories;
ANALYZE warehouses;

-- 3. CREATE MATERIALIZED VIEWS FOR ANALYTICS
-- ============================================

-- Dashboard analytics materialized view
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_dashboard_analytics AS
SELECT 
    tenant_id,
    COUNT(DISTINCT o.id) as order_count,
    COALESCE(SUM(o.total_amount), 0) as total_sales,
    COALESCE(SUM(i.quantity * p.unit_price), 0) as total_stock_value,
    COUNT(DISTINCT CASE WHEN i.quantity <= i.minimum_level THEN i.id END) as low_stock_items,
    NOW() as last_updated
FROM orders o
LEFT JOIN inventory i ON i.tenant_id = o.tenant_id
LEFT JOIN products p ON p.id = i.product_id
WHERE o.created_at >= NOW() - INTERVAL '30 days'
GROUP BY tenant_id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_dashboard_analytics_tenant ON mv_dashboard_analytics(tenant_id);

-- Product sales analytics materialized view
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_product_sales AS
SELECT 
    p.tenant_id,
    p.id as product_id,
    p.name as product_name,
    COUNT(DISTINCT oi.order_id) as order_count,
    SUM(oi.quantity) as units_sold,
    SUM(oi.quantity * oi.unit_price) as total_sales,
    MAX(o.created_at) as last_sale_date
FROM products p
LEFT JOIN order_items oi ON oi.product_id = p.id
LEFT JOIN orders o ON o.id = oi.order_id
WHERE o.created_at >= NOW() - INTERVAL '90 days'
GROUP BY p.tenant_id, p.id, p.name;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_product_sales_product ON mv_product_sales(tenant_id, product_id);
CREATE INDEX IF NOT EXISTS idx_mv_product_sales_total ON mv_product_sales(tenant_id, total_sales DESC);

-- 4. CREATE REFRESH FUNCTION FOR MATERIALIZED VIEWS
-- ============================================
CREATE OR REPLACE FUNCTION refresh_analytics_views()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_dashboard_analytics;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_product_sales;
END;
$$ LANGUAGE plpgsql;

-- 5. VACUUM AND REINDEX
-- ============================================
-- Run these periodically (e.g., weekly)
-- VACUUM ANALYZE products;
-- VACUUM ANALYZE inventory;
-- VACUUM ANALYZE orders;
-- VACUUM ANALYZE invoices;
-- REINDEX TABLE CONCURRENTLY products;
-- REINDEX TABLE CONCURRENTLY inventory;

-- 6. QUERY OPTIMIZATION SETTINGS
-- ============================================
-- Add these to postgresql.conf for better performance

-- Memory settings (adjust based on available RAM)
-- shared_buffers = 256MB              # 25% of RAM
-- effective_cache_size = 1GB          # 50-75% of RAM
-- work_mem = 16MB                     # Per operation
-- maintenance_work_mem = 128MB        # For VACUUM, CREATE INDEX

-- Query planner settings
-- random_page_cost = 1.1              # For SSD storage
-- effective_io_concurrency = 200      # For SSD storage
-- default_statistics_target = 100     # Better query plans

-- Connection settings
-- max_connections = 100
-- shared_preload_libraries = 'pg_stat_statements'

-- 7. ENABLE QUERY STATISTICS
-- ============================================
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- View slow queries
-- SELECT query, calls, total_exec_time, mean_exec_time, max_exec_time
-- FROM pg_stat_statements
-- ORDER BY mean_exec_time DESC
-- LIMIT 20;

-- 8. PARTITIONING FOR LARGE TABLES (Optional)
-- ============================================
-- For audit_logs table if it grows very large

-- CREATE TABLE audit_logs_partitioned (
--     LIKE audit_logs INCLUDING ALL
-- ) PARTITION BY RANGE (created_at);

-- CREATE TABLE audit_logs_2025_01 PARTITION OF audit_logs_partitioned
--     FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

-- 9. MONITORING QUERIES
-- ============================================

-- Check index usage
-- SELECT 
--     schemaname,
--     tablename,
--     indexname,
--     idx_scan,
--     idx_tup_read,
--     idx_tup_fetch
-- FROM pg_stat_user_indexes
-- WHERE idx_scan = 0
-- ORDER BY schemaname, tablename;

-- Check table sizes
-- SELECT 
--     schemaname,
--     tablename,
--     pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
-- FROM pg_tables
-- WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
-- ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Check cache hit ratio (should be > 99%)
-- SELECT 
--     sum(heap_blks_read) as heap_read,
--     sum(heap_blks_hit) as heap_hit,
--     sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) as ratio
-- FROM pg_statio_user_tables;

-- 10. CLEANUP OLD DATA (Optional)
-- ============================================
-- Archive old audit logs (older than 90 days)
-- DELETE FROM audit_logs WHERE created_at < NOW() - INTERVAL '90 days';

-- Archive old stock adjustments (older than 180 days)
-- DELETE FROM stock_adjustments WHERE adjusted_at < NOW() - INTERVAL '180 days';

COMMIT;
