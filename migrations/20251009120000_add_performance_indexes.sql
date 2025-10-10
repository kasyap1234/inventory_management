-- +goose Up
-- Adds missing indexes to improve query performance for critical workflows

BEGIN;

-- Inventory lookups by tenant/product/warehouse
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_product
    ON inventory (tenant_id, product_id);

CREATE INDEX IF NOT EXISTS idx_inventory_tenant_warehouse_product
    ON inventory (tenant_id, warehouse_id, product_id);

-- Order workflow filtering by tenant and status
CREATE INDEX IF NOT EXISTS idx_orders_tenant_status
    ON orders (tenant_id, status, order_type);

-- Invoice lookups by tenant and status/due date
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_status_due
    ON invoices (tenant_id, status, due_date);

-- Audit log history lookups by tenant/table/record
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_table_record
    ON audit_logs (tenant_id, table_name, record_id);

-- Audit log chronological queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_created_at
    ON audit_logs (tenant_id, created_at DESC);

COMMIT;

-- +goose Down

BEGIN;

DROP INDEX IF EXISTS idx_audit_logs_tenant_created_at;
DROP INDEX IF EXISTS idx_audit_logs_tenant_table_record;
DROP INDEX IF EXISTS idx_invoices_tenant_status_due;
DROP INDEX IF EXISTS idx_orders_tenant_status;
DROP INDEX IF EXISTS idx_inventory_tenant_warehouse_product;
DROP INDEX IF EXISTS idx_inventory_tenant_product;

COMMIT;
