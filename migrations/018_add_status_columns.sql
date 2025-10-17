-- Add status columns to suppliers, distributors, and warehouses tables
-- This migration ensures these tables have status columns needed for status-based queries and indexes

-- Add status column to suppliers table
ALTER TABLE IF EXISTS suppliers ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended', 'archived'));

-- Add status column to distributors table
ALTER TABLE IF EXISTS distributors ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended', 'archived'));

-- Add status column to warehouses table
ALTER TABLE IF EXISTS warehouses ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended', 'archived'));

-- Create indexes for the new status columns
CREATE INDEX IF NOT EXISTS idx_suppliers_tenant_status ON suppliers (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_distributors_tenant_status ON distributors (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_warehouses_tenant_status ON warehouses (tenant_id, status);