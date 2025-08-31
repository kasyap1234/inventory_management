-- Enterprise Search & Filtering Optimization Indexes
-- Migration: 20240130000001_optimize_search_indexes.sql

-- ============================================================================
-- PRODUCTS SEARCH OPTIMIZATION INDEXES
-- ============================================================================

-- Composite index for advanced product search with tenant isolation
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_advanced_search
ON products (tenant_id, created_at DESC)
WHERE tenant_id IS NOT NULL;

-- Full-text search indexes for products
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_search_name
ON products USING GIN (to_tsvector('english', COALESCE(name, '')));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_search_description
ON products USING GIN (to_tsvector('english', COALESCE(description, '')));

-- Composite index for category-based searches
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_category_search
ON products (tenant_id, category_id, created_at DESC)
WHERE category_id IS NOT NULL;

-- Index for barcode lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_barcode
ON products (tenant_id, barcode)
WHERE barcode IS NOT NULL AND barcode != '';

-- Index for price range queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_price_range
ON products (tenant_id, unit_price)
WHERE unit_price IS NOT NULL;

-- Index for expiry date queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_expiry_search
ON products (tenant_id, expiry_date)
WHERE expiry_date IS NOT NULL;

-- Index for quantity filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_quantity_search
ON products (tenant_id, quantity DESC)
WHERE quantity IS NOT NULL;

-- ============================================================================
-- ORDERS SEARCH OPTIMIZATION INDEXES
-- ============================================================================

-- Composite index for advanced order search
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_advanced_search
ON orders (tenant_id, order_date DESC)
WHERE tenant_id IS NOT NULL;

-- Index for order date range queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_date_range
ON orders (tenant_id, order_date)
WHERE order_date IS NOT NULL;

-- Index for expected delivery date queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_delivery_search
ON orders (tenant_id, expected_delivery)
WHERE expected_delivery IS NOT NULL;

-- Index for order type and status filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_type_status
ON orders (tenant_id, order_type, status);

-- Full-text search on order notes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_notes_search
ON orders USING GIN (to_tsvector('english', COALESCE(notes, '')));

-- Index for value range queries (quantity * unit_price)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_value_composite
ON orders (quantity, unit_price);

-- Index for supplier filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_supplier
ON orders (tenant_id, supplier_id)
WHERE supplier_id IS NOT NULL;

-- Index for distributor filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_distributor
ON orders (tenant_id, distributor_id)
WHERE distributor_id IS NOT NULL;

-- ============================================================================
-- INVENTORY SEARCH OPTIMIZATION INDEXES
-- ============================================================================

-- Composite index for inventory search
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inventory_search
ON inventory (tenant_id, last_updated DESC)
WHERE tenant_id IS NOT NULL;

-- Index for warehouse filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inventory_warehouse
ON inventory (tenant_id, warehouse_id);

-- Index for product filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inventory_product
ON inventory (tenant_id, product_id);

-- Index for quantity filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inventory_quantity
ON inventory (tenant_id, quantity DESC);

-- Composite index for warehouse and product lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inventory_warehouse_product
ON inventory (tenant_id, warehouse_id, product_id);

-- ============================================================================
-- CROSS-TABLE SEARCH OPTIMIZATION INDEXES
-- ============================================================================

-- Indexes to support full-text search across related tables

-- Product name search to support order searches
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_name_gin
ON products USING GIN (to_tsvector('english', COALESCE(name, '')))
WHERE tenant_id IS NOT NULL;

-- Supplier name for order searches
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_suppliers_name_search
ON suppliers (tenant_id, name)
WHERE name IS NOT NULL;

-- Distributor name for order searches
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_distributors_name_search
ON distributors (tenant_id, name)
WHERE name IS NOT NULL;

-- Warehouse name for inventory searches
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_warehouses_name_search
ON warehouses (tenant_id, name)
WHERE name IS NOT NULL;

-- Category name for product searches
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_categories_name_search
ON categories (tenant_id, name)
WHERE name IS NOT NULL;

-- ============================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- ============================================================================

-- Index for bulk operation queries (by last updated)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_bulk_operations
ON products (tenant_id, updated_at DESC);

-- Index for bulk order status updates
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_bulk_status
ON orders (tenant_id, status, created_at DESC);

-- Index for bulk inventory updates
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inventory_bulk_updates
ON inventory (tenant_id, last_updated DESC);

-- ============================================================================
-- ANALYTICS OPTIMIZATION INDEXES
-- ============================================================================

-- Composite index for search analytics queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_search_analytics_composite
ON search_analytics (tenant_id, timestamp DESC)
WHERE tenant_id IS NOT NULL;

-- Index for search term frequency analysis
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_search_analytics_terms
ON search_analytics (search_term, timestamp DESC);

-- ============================================================================
-- CLEANUP: Remove any unused legacy indexes
-- ============================================================================

-- Remove legacy single-column indexes if they exist and are not used
DROP INDEX IF EXISTS products_name_idx;
DROP INDEX IF EXISTS products_created_at_idx;
DROP INDEX IF EXISTS orders_order_date_idx;

-- ============================================================================
-- MAINTENANCE OPTIMIZATION
-- ============================================================================

-- Analyze all tables after index creation for optimal query planning
ANALYZE products;
ANALYZE orders;
ANALYZE inventory;
ANALYZE categories;
ANALYZE suppliers;
ANALYZE distributors;
ANALYZE warehouses;

-- ============================================================================
-- COMMENT EXPLANATION
-- ============================================================================

COMMENT ON INDEX idx_products_advanced_search IS 'Optimizes advanced product search queries with tenant isolation and sorting by creation date';
COMMENT ON INDEX idx_orders_advanced_search IS 'Optimizes advanced order search queries with tenant isolation and date-based sorting';
COMMENT ON INDEX idx_inventory_search IS 'Optimizes inventory search queries with tenant isolation and recency-based sorting';

COMMENT ON INDEX idx_products_search_name IS 'Full-text search index for product names to support fuzzy matching';
COMMENT ON INDEX idx_products_search_description IS 'Full-text search index for product descriptions';
COMMENT ON INDEX idx_orders_notes_search IS 'Full-text search index for order notes';

-- ============================================================================
-- POST-MIGRATION VALIDATION
-- ============================================================================

-- Example query to test index effectiveness:
/*
EXPLAIN ANALYZE
SELECT p.* FROM products p
WHERE p.tenant_id = '550e8400-e29b-41d4-a716-446655440000'
  AND (to_tsvector('english', COALESCE(p.name, '')) @@ plainto_tsquery('english', 'fertilizer'))
  AND p.unit_price BETWEEN 100 AND 1000
  AND p.quantity > 10
  AND p.category_id = '550e8400-e29b-41d4-a716-446655440001'
ORDER BY p.created_at DESC
LIMIT 50 OFFSET 0;
*/