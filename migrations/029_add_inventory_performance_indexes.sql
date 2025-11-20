-- Migration: Add performance indexes for inventory queries
-- Description: Adds indexes to optimize inventory list and search queries with JOINs

-- Index for inventory lookups by tenant and warehouse/product
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_warehouse_product 
ON inventory(tenant_id, warehouse_id, product_id);

-- Index for inventory sorting by last_updated
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_last_updated 
ON inventory(tenant_id, last_updated DESC);

-- Index for product name searches (case-insensitive)
CREATE INDEX IF NOT EXISTS idx_products_tenant_name 
ON products(tenant_id, LOWER(name));

-- Index for warehouse name searches (case-insensitive)
CREATE INDEX IF NOT EXISTS idx_warehouses_tenant_name 
ON warehouses(tenant_id, LOWER(name));

-- Index for product barcode lookups
CREATE INDEX IF NOT EXISTS idx_products_tenant_barcode 
ON products(tenant_id, barcode) WHERE barcode IS NOT NULL;

-- Index for inventory quantity filtering
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_quantity 
ON inventory(tenant_id, quantity);
