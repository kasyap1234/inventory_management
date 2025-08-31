-- Table Partitioning Implementation for Large Datasets
-- Migration: 20240130000002_table_partitioning.sql

-- ============================================================================
-- AUDIT_LOGS TABLE PARTITIONING
-- ============================================================================

-- Partition audit_logs table by tenant_id using hash partitioning for even distribution
-- This enables better performance for multi-tenant scenarios with frequent audit logging

DROP TABLE IF EXISTS audit_logs CASCADE;

-- Create partitioned master table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    table_name VARCHAR(50) NOT NULL,
    record_id VARCHAR(100) NOT NULL,
    action VARCHAR(20) NOT NULL,
    new_values JSONB,
    old_values JSONB,
    changed_by UUID REFERENCES users(id),
    deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
) PARTITION BY HASH (tenant_id);

-- Create 8 hash partitions for optimal distribution
CREATE TABLE audit_logs_p0 PARTITION OF audit_logs FOR VALUES WITH (modulus 8, remainder 0);
CREATE TABLE audit_logs_p1 PARTITION OF audit_logs FOR VALUES WITH (modulus 8, remainder 1);
CREATE TABLE audit_logs_p2 PARTITION OF audit_logs FOR VALUES WITH (modulus 8, remainder 2);
CREATE TABLE audit_logs_p3 PARTITION OF audit_logs FOR VALUES WITH (modulus 8, remainder 3);
CREATE TABLE audit_logs_p4 PARTITION OF audit_logs FOR VALUES WITH (modulus 8, remainder 4);
CREATE TABLE audit_logs_p5 PARTITION OF audit_logs FOR VALUES WITH (modulus 8, remainder 5);
CREATE TABLE audit_logs_p6 PARTITION OF audit_logs FOR VALUES WITH (modulus 8, remainder 6);
CREATE TABLE audit_logs_p7 PARTITION OF audit_logs FOR VALUES WITH (modulus 8, remainder 7);

-- Create indexes on partitioned tables after partitioning
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_tenant_created
ON audit_logs (tenant_id, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_tenant_table_record
ON audit_logs (tenant_id, table_name, record_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_tenant_changed_by
ON audit_logs (tenant_id, changed_by) WHERE changed_by IS NOT NULL;

-- ============================================================================
-- ORDERS TABLE PARTITIONING (RANGE PARTITIONING BY DATE)
-- ============================================================================

-- Create a backup of existing orders data before partitioning
CREATE TABLE orders_backup AS SELECT * FROM orders;

-- Recreate orders table with partitioning
DROP TABLE IF EXISTS orders CASCADE;

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    order_type VARCHAR(10) NOT NULL CHECK (order_type IN ('purchase', 'sales')),
    supplier_id UUID REFERENCES suppliers(id),
    distributor_id UUID REFERENCES distributors(id),
    product_id UUID NOT NULL REFERENCES products(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending' NOT NULL CHECK (status IN ('pending', 'approved', 'received', 'shipped', 'delivered', 'cancelled')),
    order_date DATE DEFAULT NOW(),
    expected_delivery DATE,
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CHECK (order_type = 'purchase' AND supplier_id IS NOT NULL AND distributor_id IS NULL OR
        order_type = 'sales' AND distributor_id IS NOT NULL AND supplier_id IS NULL)
) PARTITION BY RANGE (order_date);

-- Create monthly partitions for orders (current and next 12 months)
-- Generate partitions dynamically based on current date
DO $$
DECLARE
    current_year INT := EXTRACT(YEAR FROM CURRENT_DATE);
    current_month INT := EXTRACT(MONTH FROM CURRENT_DATE);
    partition_year INT;
    partition_month INT;
    partition_date DATE;
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
    i INT := 0;
BEGIN
    WHILE i < 24 LOOP  -- Create 24 months of partitions
        partition_year := current_year + ((current_month - 1 + i) / 12);
        partition_month := ((current_month - 1 + i) % 12) + 1;

        partition_date := DATE (partition_year || '-' || partition_month || '-01');
        start_date := partition_date;
        end_date := partition_date + INTERVAL '1 month';

        partition_name := 'orders_y' || partition_year || '_m' || LPAD(partition_month::TEXT, 2, '0');

        -- Create partition
        EXECUTE format('CREATE TABLE %I PARTITION OF orders FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date);

        i := i + 1;
    END LOOP;
END $$;

-- Add a default partition for any future dates beyond our defined ranges
CREATE TABLE orders_default PARTITION OF orders DEFAULT;

-- Restore data from backup
INSERT INTO orders (id, tenant_id, order_type, supplier_id, distributor_id, product_id,
                   warehouse_id, quantity, unit_price, status, order_date, expected_delivery,
                   notes, created_at, updated_at)
SELECT id, tenant_id, order_type, supplier_id, distributor_id, product_id,
       warehouse_id, quantity, unit_price, status, order_date, expected_delivery,
       notes, created_at, updated_at
FROM orders_backup;

-- Drop backup table after successful restoration
DROP TABLE orders_backup;

-- ============================================================================
-- INVENTORY TABLE PARTITIONING (RANGE PARTITIONING BY LAST_UPDATED)
-- ============================================================================

-- Create backup of existing inventory data
CREATE TABLE inventory_backup AS SELECT * FROM inventory;

-- Recreate inventory table with partitioning based on last_updated timestamp
DROP TABLE IF EXISTS inventory CASCADE;

CREATE TABLE inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    last_updated TIMESTAMP DEFAULT NOW(),
    UNIQUE (tenant_id, warehouse_id, product_id)
) PARTITION BY RANGE (last_updated);

-- Create partitions by quarter for inventory updates
DO $$
DECLARE
    current_year INT := EXTRACT(YEAR FROM CURRENT_DATE);
    partition_year INT;
    quarters TEXT[] := ARRAY['01-01', '04-01', '07-01', '10-01'];
    quarter_names TEXT[] := ARRAY['_q1', '_q2', '_q3', '_q4'];
    i INT := 0;
    j INT;
BEGIN
    FOR j IN 1..4 LOOP  -- 4 quarters per year
        partition_year := current_year + i;
        PERFORM create_inventory_partition(partition_year, j);
    END LOOP;
END $$;

-- Function to create inventory partitions
CREATE OR REPLACE FUNCTION create_inventory_partition(partition_year INT, quarter_num INT)
RETURNS VOID AS $$
DECLARE
    quarters TEXT[] := ARRAY['01-01', '04-01', '07-01', '10-01'];
    start_date DATE := DATE (partition_year || '-' || quarters[quarter_num]);
    end_date DATE;
    partition_name TEXT := 'inventory_' || partition_year || '_q' || quarter_num;
BEGIN
    IF quarter_num < 4 THEN
        end_date := DATE (partition_year || '-' || quarters[quarter_num + 1]);
    ELSE
        end_date := DATE ((partition_year + 1) || '-' || quarters[1]);
    END IF;

    EXECUTE format('CREATE TABLE %I PARTITION OF inventory FOR VALUES FROM (%L) TO (%L)',
        partition_name, start_date, end_date);
END;
$$ LANGUAGE plpgsql;

-- Create initial inventory partitions
SELECT create_inventory_partition(EXTRACT(YEAR FROM CURRENT_DATE)::INT, 1);
SELECT create_inventory_partition(EXTRACT(YEAR FROM CURRENT_DATE)::INT, 2);
SELECT create_inventory_partition(EXTRACT(YEAR FROM CURRENT_DATE)::INT, 3);
SELECT create_inventory_partition(EXTRACT(YEAR FROM CURRENT_DATE)::INT, 4);

-- Default partition
CREATE TABLE inventory_default PARTITION OF inventory DEFAULT;

-- Restore data from backup
INSERT INTO inventory (id, tenant_id, warehouse_id, product_id, quantity, last_updated)
SELECT id, tenant_id, warehouse_id, product_id, quantity, last_updated
FROM inventory_backup;

-- Drop backup table
DROP TABLE inventory_backup;

-- ============================================================================
-- OPTIMIZATION INDEXES FOR PARTITIONED TABLES
-- ============================================================================

-- Partition-aware indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_partition_aware
ON orders (tenant_id, order_date DESC, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inventory_partition_aware
ON inventory (tenant_id, last_updated DESC, quantity);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_partition_aware
ON audit_logs (tenant_id, created_at DESC, table_name);

-- ============================================================================
-- PARTITION MAINTENANCE PROCEDURES
-- ============================================================================

-- Function to create future partitions automatically
CREATE OR REPLACE FUNCTION create_future_partitions()
RETURNS VOID AS $$
DECLARE
    max_order_date DATE;
    max_inventory_date DATE;
BEGIN
    -- Create future order partitions (next 6 months)
    SELECT MAX(order_date) INTO max_order_date FROM orders;
    IF max_order_date IS NOT NULL THEN
        -- Logic to create partitions if approaching the end of current range
        PERFORM create_order_partitions_future(max_order_date);
    END IF;

    -- Create future inventory partitions (next 2 quarters)
    SELECT MAX(last_updated) INTO max_inventory_date FROM inventory;
    IF max_inventory_date IS NOT NULL THEN
        -- Logic to create inventory partitions if needed
        PERFORM create_inventory_partitions_future(max_inventory_date);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- AUTO-VACUUM OPTIMIZATION FOR PARTITIONED TABLES
-- ============================================================================

-- Set aggressive autovacuum settings for partitioned tables
ALTER TABLE audit_logs SET (autovacuum_vacuum_scale_factor = 0.1);
ALTER TABLE audit_logs SET (autovacuum_analyze_scale_factor = 0.05);

ALTER TABLE orders SET (autovacuum_vacuum_scale_factor = 0.1);
ALTER TABLE orders SET (autovacuum_analyze_scale_factor = 0.05);

ALTER TABLE inventory SET (autovacuum_vacuum_scale_factor = 0.1);
ALTER TABLE inventory SET (autovacuum_analyze_scale_factor = 0.05);

-- ============================================================================
-- MONITORING QUERIES FOR PARTITIONED TABLES
-- ============================================================================

-- Query to check partition sizes
/*
SELECT
    schemaname,
    tablename,
    n_tup_ins,
    n_tup_upd,
    n_tup_del,
    n_live_tup,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_stat_user_tables
WHERE tablename LIKE '%audit_logs%' OR tablename LIKE '%orders%' OR tablename LIKE '%inventory%'
ORDER BY n_live_tup DESC;
*/

-- Query to check partition distribution
/*
SELECT
    s.schemaname,
    s.tablename,
    s.n_live_tup,
    pg_size_pretty(pg_total_relation_size(s.schemaname||'.'||s.tablename)) as size,
    CASE
        WHEN s.tablename ~ '_p[0-7]$' THEN 'Audit Partition'
        WHEN s.tablename ~ '_y[0-9]{4}_m[0-9]{2}$' THEN 'Order Partition'
        WHEN s.tablename ~ '_[0-9]{4}_q[1-4]$' THEN 'Inventory Partition'
        ELSE 'Default'
    END as partition_type
FROM pg_stat_user_tables s
WHERE s.tablename LIKE '%audit_logs%'
   OR s.tablename LIKE '%orders%'
   OR s.tablename LIKE '%inventory%'
ORDER BY s.schemaname, s.tablename;
*/

-- ============================================================================
-- CLEANUP
-- ============================================================================

-- Drop helper functions after partition creation
DROP FUNCTION IF EXISTS create_inventory_partition(INT, INT);
DROP FUNCTION IF EXISTS create_order_partitions_future(DATE);
DROP FUNCTION IF EXISTS create_inventory_partitions_future(TIMESTAMP);