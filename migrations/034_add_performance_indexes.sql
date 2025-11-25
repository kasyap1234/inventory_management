-- Migration: 034_add_performance_indexes.sql
-- Description: Add performance indexes for frequently queried columns
-- Created: 2025-11-25
-- 
-- These indexes optimize common query patterns identified during code analysis:
-- 1. Order status filtering (used in order list views and dashboards)
-- 2. Order date range queries (used in reports and analytics)
-- 3. Inventory lookups by product and warehouse combination
-- 4. Tenant-scoped order queries (most common access pattern)

-- Index for order status filtering
-- Speeds up: GET /orders?status=pending, dashboard status counts
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);

-- Index for order date range queries
-- Speeds up: Date range filters in order reports, analytics
CREATE INDEX IF NOT EXISTS idx_orders_order_date ON orders(order_date);

-- Composite index for inventory lookups
-- Speeds up: Inventory checks during order creation, stock updates
CREATE INDEX IF NOT EXISTS idx_inventory_product_warehouse ON inventory(product_id, warehouse_id);

-- Composite index for tenant-scoped order queries
-- Speeds up: All order list operations (most frequent query pattern)
CREATE INDEX IF NOT EXISTS idx_orders_tenant_status ON orders(tenant_id, status);

-- Composite index for tenant-scoped order date queries
-- Speeds up: Order date range filters with tenant isolation
CREATE INDEX IF NOT EXISTS idx_orders_tenant_order_date ON orders(tenant_id, order_date);

-- Index for reserved quantity lookups
-- Speeds up: Stock availability checks, reservation operations
CREATE INDEX IF NOT EXISTS idx_inventory_reserved_quantity ON inventory(tenant_id, product_id, reserved_quantity) 
WHERE reserved_quantity > 0;

-- Index for low stock alerts
-- Speeds up: Low stock detection queries, alert generation
CREATE INDEX IF NOT EXISTS idx_inventory_quantity ON inventory(tenant_id, quantity);

-- Index for product search in inventory
-- Speeds up: Product-based inventory queries
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_product ON inventory(tenant_id, product_id);

-- Partial index for active reservations only
-- Speeds up: Active reservation lookups (ignores historical data)
CREATE INDEX IF NOT EXISTS idx_inventory_reservations_active 
ON inventory_reservations(tenant_id, product_id, status) 
WHERE status = 'active';

-- Index for order items by order
-- Speeds up: Order detail retrieval
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);

-- Index for audit logs by table and record
-- Speeds up: Entity history lookups
CREATE INDEX IF NOT EXISTS idx_audit_logs_table_record 
ON audit_logs(tenant_id, table_name, record_id);

-- Verify indexes were created
DO $$
BEGIN
    RAISE NOTICE 'Performance indexes migration completed successfully';
END $$;
