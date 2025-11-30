-- Migration: Add cursor pagination indexes for performance
-- These composite indexes optimize cursor-based pagination queries that use (created_at, id) ordering

-- Products cursor pagination index
-- Optimizes: WHERE tenant_id = $1 AND (created_at, id) < ($2, $3) ORDER BY created_at DESC, id DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_cursor_pagination 
ON products (tenant_id, created_at DESC, id DESC);

-- Orders cursor pagination index
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_cursor_pagination 
ON orders (tenant_id, created_at DESC, id DESC);

-- Inventory cursor pagination index
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inventory_cursor_pagination 
ON inventory (tenant_id, last_updated DESC, id DESC);

-- Invoices cursor pagination index  
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_cursor_pagination 
ON invoices (tenant_id, created_at DESC, id DESC);

-- Audit logs cursor pagination index (commonly used for recent activity)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_cursor_pagination 
ON audit_logs (tenant_id, created_at DESC, id DESC);

-- Index for analytics aggregates on invoices (total_amount, gst columns)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_analytics_aggregates 
ON invoices (tenant_id) INCLUDE (total_amount, cgst, sgst, igst);

-- Index for inventory analytics (joins with products for stock value)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inventory_analytics 
ON inventory (tenant_id, product_id) INCLUDE (quantity);

-- Comment: CONCURRENTLY allows index creation without blocking table writes
-- This is important for production systems with high write loads
