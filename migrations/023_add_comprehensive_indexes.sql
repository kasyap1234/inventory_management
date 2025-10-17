-- Add performance indexes for commonly queried columns
-- Migration: 024_add_performance_indexes.sql
-- This migration adds indexes to improve query performance

-- Users table indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_tenant_email ON users(tenant_id, email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- Products table indexes
CREATE INDEX IF NOT EXISTS idx_products_tenant_category ON products(tenant_id, category_id);
CREATE INDEX IF NOT EXISTS idx_products_name_search ON products USING gin(to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_products_barcode ON products(barcode) WHERE barcode IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at DESC);

-- Orders table indexes
CREATE INDEX IF NOT EXISTS idx_orders_tenant_status ON orders(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_orders_tenant_created ON orders(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);

-- Order items table indexes
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);

-- Invoices table indexes
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_status ON invoices(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_invoices_order_id ON invoices(order_id);
CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(due_date) WHERE status != 'paid';
CREATE INDEX IF NOT EXISTS idx_invoices_created_at ON invoices(created_at DESC);

-- Inventory table indexes  
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_warehouse ON inventory(tenant_id, warehouse_id);
CREATE INDEX IF NOT EXISTS idx_inventory_product_id ON inventory(product_id);
CREATE INDEX IF NOT EXISTS idx_inventory_low_stock ON inventory(quantity) WHERE quantity <= reorder_level;

-- Audit logs table indexes (for performance of history queries)
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_table ON audit_logs(tenant_id, table_name);
CREATE INDEX IF NOT EXISTS idx_audit_logs_record_id ON audit_logs(record_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);

-- User roles indexes
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_tenant ON user_roles(tenant_id, user_id);

-- Role permissions indexes
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Notifications table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_notifications_tenant_user ON notifications(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);

-- Subscriptions table indexes
CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant_id ON subscriptions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_next_billing ON subscriptions(next_billing_date) WHERE status = 'active';

-- Warehouses table indexes
CREATE INDEX IF NOT EXISTS idx_warehouses_tenant_id ON warehouses(tenant_id);
CREATE INDEX IF NOT EXISTS idx_warehouses_status ON warehouses(status);

-- Suppliers table indexes
CREATE INDEX IF NOT EXISTS idx_suppliers_tenant_id ON suppliers(tenant_id);

-- Distributors table indexes
CREATE INDEX IF NOT EXISTS idx_distributors_tenant_id ON distributors(tenant_id);

-- Categories table indexes
CREATE INDEX IF NOT EXISTS idx_categories_tenant_id ON categories(tenant_id);
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id) WHERE parent_id IS NOT NULL;

-- Product images indexes
CREATE INDEX IF NOT EXISTS idx_product_images_product_id ON product_images(product_id);
CREATE INDEX IF NOT EXISTS idx_product_images_tenant_id ON product_images(tenant_id);

-- Add composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_products_tenant_status ON products(tenant_id, deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_orders_tenant_dates ON orders(tenant_id, created_at DESC, status);
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_dates ON invoices(tenant_id, invoice_date DESC, status);

-- Add partial indexes for filtered queries
CREATE INDEX IF NOT EXISTS idx_orders_pending ON orders(tenant_id, created_at DESC) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_orders_approved ON orders(tenant_id, created_at DESC) WHERE status = 'approved';
CREATE INDEX IF NOT EXISTS idx_invoices_unpaid ON invoices(tenant_id, due_date) WHERE status IN ('pending', 'overdue');

-- Add GIN index for JSON columns if they exist
-- This is useful for notification_data or metadata columns
CREATE INDEX IF NOT EXISTS idx_notifications_data ON notifications USING gin(notification_data) WHERE notification_data IS NOT NULL;

-- Add indexes for timestamp columns used in analytics
CREATE INDEX IF NOT EXISTS idx_orders_completed_at ON orders(completed_at DESC) WHERE completed_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_invoices_paid_at ON invoices(paid_at DESC) WHERE paid_at IS NOT NULL;

-- Analyze tables after index creation to update statistics
ANALYZE users;
ANALYZE products;
ANALYZE orders;
ANALYZE order_items;
ANALYZE invoices;
ANALYZE inventory;
ANALYZE audit_logs;
ANALYZE user_roles;
ANALYZE role_permissions;
ANALYZE subscriptions;
ANALYZE warehouses;
ANALYZE suppliers;
ANALYZE distributors;
ANALYZE categories;
ANALYZE product_images;

-- Note: Some indexes may fail if columns don't exist, which is fine
-- The IF NOT EXISTS clause prevents errors on re-runs
