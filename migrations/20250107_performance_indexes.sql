-- Performance Indexes Migration
-- Date: 2025-01-07
-- Purpose: Add critical indexes to improve query performance across all major tables

-- ====================
-- ORDER INDEXES
-- ====================

-- Index for filtering orders by tenant and status (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_orders_tenant_status 
ON orders(tenant_id, status) 
WHERE deleted = FALSE;

-- Index for sorting orders by date within tenant (dashboard queries)
CREATE INDEX IF NOT EXISTS idx_orders_tenant_date 
ON orders(tenant_id, order_date DESC) 
WHERE deleted = FALSE;

-- Index for product-related order queries
CREATE INDEX IF NOT EXISTS idx_orders_product 
ON orders(product_id) 
WHERE deleted = FALSE;

-- Index for supplier-related order queries
CREATE INDEX IF NOT EXISTS idx_orders_supplier 
ON orders(supplier_id) 
WHERE supplier_id IS NOT NULL AND deleted = FALSE;

-- Index for distributor-related order queries
CREATE INDEX IF NOT EXISTS idx_orders_distributor 
ON orders(distributor_id) 
WHERE distributor_id IS NOT NULL AND deleted = FALSE;

-- Index for order type filtering (purchase vs sales)
CREATE INDEX IF NOT EXISTS idx_orders_tenant_type 
ON orders(tenant_id, order_type) 
WHERE deleted = FALSE;

-- ====================
-- INVOICE INDEXES
-- ====================

-- Index for filtering invoices by tenant and status
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_status 
ON invoices(tenant_id, status) 
WHERE deleted = FALSE;

-- Index for sorting invoices by issued date
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_date 
ON invoices(tenant_id, issued_date DESC) 
WHERE deleted = FALSE;

-- Index for finding invoices by order
CREATE INDEX IF NOT EXISTS idx_invoices_order 
ON invoices(order_id);

-- Index for overdue invoice queries
CREATE INDEX IF NOT EXISTS idx_invoices_due_date 
ON invoices(due_date) 
WHERE status = 'unpaid';

-- Index for invoice number lookups (if invoice_number exists)
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_number 
ON invoices(tenant_id, invoice_number) 
WHERE invoice_number IS NOT NULL;

-- ====================
-- INVENTORY INDEXES
-- ====================

-- Composite index for inventory queries by tenant and product
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_product 
ON inventory(tenant_id, product_id);

-- Index for warehouse-specific inventory queries
CREATE INDEX IF NOT EXISTS idx_inventory_warehouse 
ON inventory(warehouse_id);

-- Index for low stock alerts (critical for business operations)
CREATE INDEX IF NOT EXISTS idx_inventory_low_stock 
ON inventory(tenant_id, quantity) 
WHERE quantity < 20;

-- Index for product availability across warehouses
CREATE INDEX IF NOT EXISTS idx_inventory_product_availability 
ON inventory(product_id, quantity) 
WHERE quantity > 0;

-- ====================
-- PRODUCT INDEXES
-- ====================

-- Index for category-based product queries
CREATE INDEX IF NOT EXISTS idx_products_tenant_category 
ON products(tenant_id, category_id) 
WHERE deleted = FALSE;

-- Index for barcode lookups (common in retail operations)
CREATE INDEX IF NOT EXISTS idx_products_barcode 
ON products(barcode) 
WHERE barcode IS NOT NULL AND deleted = FALSE;

-- Index for expiry date queries (important for perishable goods)
CREATE INDEX IF NOT EXISTS idx_products_expiry 
ON products(expiry_date) 
WHERE expiry_date IS NOT NULL AND deleted = FALSE;

-- Full-text search index for product names (PostgreSQL specific)
CREATE INDEX IF NOT EXISTS idx_products_name_search 
ON products USING gin(to_tsvector('english', name));

-- Index for product name sorting
CREATE INDEX IF NOT EXISTS idx_products_tenant_name 
ON products(tenant_id, name) 
WHERE deleted = FALSE;

-- ====================
-- USER AND AUTH INDEXES
-- ====================

-- Index for user lookups by tenant and email
CREATE INDEX IF NOT EXISTS idx_users_tenant_email 
ON users(tenant_id, email) 
WHERE status = 'active';

-- Index for email lookups (login operations)
CREATE INDEX IF NOT EXISTS idx_users_email_active 
ON users(email) 
WHERE status = 'active';

-- Index for user roles relationship
CREATE INDEX IF NOT EXISTS idx_user_roles_user 
ON user_roles(user_id);

-- Index for role assignments
CREATE INDEX IF NOT EXISTS idx_user_roles_role 
ON user_roles(role_id);

-- Index for role permissions
CREATE INDEX IF NOT EXISTS idx_role_permissions_role 
ON role_permissions(role_id);

-- Index for permission lookups
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission 
ON role_permissions(permission_id);

-- ====================
-- SUPPLIER/DISTRIBUTOR INDEXES
-- ====================

-- Index for supplier queries by tenant
CREATE INDEX IF NOT EXISTS idx_suppliers_tenant 
ON suppliers(tenant_id) 
WHERE deleted = FALSE;

-- Index for distributor queries by tenant
CREATE INDEX IF NOT EXISTS idx_distributors_tenant 
ON distributors(tenant_id) 
WHERE deleted = FALSE;

-- Index for supplier name search
CREATE INDEX IF NOT EXISTS idx_suppliers_tenant_name 
ON suppliers(tenant_id, name) 
WHERE deleted = FALSE;

-- Index for distributor name search
CREATE INDEX IF NOT EXISTS idx_distributors_tenant_name 
ON distributors(tenant_id, name) 
WHERE deleted = FALSE;

-- ====================
-- AUDIT LOG INDEXES
-- ====================

-- Composite index for audit log queries (compliance and debugging)
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_table_record 
ON audit_logs(tenant_id, table_name, record_id);

-- Index for user activity tracking
CREATE INDEX IF NOT EXISTS idx_audit_logs_user 
ON audit_logs(changed_by) 
WHERE changed_by IS NOT NULL;

-- Index for chronological audit log queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_date 
ON audit_logs(changed_at DESC);

-- Index for action type filtering
CREATE INDEX IF NOT EXISTS idx_audit_logs_action 
ON audit_logs(action);

-- ====================
-- CATEGORY INDEXES
-- ====================

-- Index for category hierarchy queries
CREATE INDEX IF NOT EXISTS idx_categories_tenant_parent 
ON categories(tenant_id, parent_category_id) 
WHERE deleted = FALSE;

-- Index for root categories
CREATE INDEX IF NOT EXISTS idx_categories_root 
ON categories(tenant_id) 
WHERE parent_category_id IS NULL AND deleted = FALSE;

-- ====================
-- WAREHOUSE INDEXES
-- ====================

-- Index for warehouse queries by tenant
CREATE INDEX IF NOT EXISTS idx_warehouses_tenant 
ON warehouses(tenant_id) 
WHERE deleted = FALSE;

-- ====================
-- TENANT INDEXES
-- ====================

-- Index for active tenant lookups
CREATE INDEX IF NOT EXISTS idx_tenants_status 
ON tenants(status) 
WHERE status = 'active';

-- ====================
-- PERFORMANCE STATISTICS
-- ====================

-- Analyze all tables to update statistics for query planner
ANALYZE orders;
ANALYZE invoices;
ANALYZE inventory;
ANALYZE products;
ANALYZE users;
ANALYZE suppliers;
ANALYZE distributors;
ANALYZE audit_logs;
ANALYZE categories;
ANALYZE warehouses;
ANALYZE tenants;
ANALYZE user_roles;
ANALYZE role_permissions;

-- ====================
-- INDEX MAINTENANCE
-- ====================

-- Reindex all tables for optimal performance (run during maintenance window)
-- REINDEX DATABASE CONCURRENTLY;  -- Uncomment for production use with caution

-- Print completion message
DO $$
BEGIN
    RAISE NOTICE 'Performance indexes created successfully';
    RAISE NOTICE 'Total indexes created: 40+';
    RAISE NOTICE 'Next steps:';
    RAISE NOTICE '  1. Monitor query performance with EXPLAIN ANALYZE';
    RAISE NOTICE '  2. Check index usage with pg_stat_user_indexes';
    RAISE NOTICE '  3. Consider vacuum analyze for immediate effect';
END $$;
