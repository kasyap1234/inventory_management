-- Migration: Create stock adjustments history table
-- Description: Adds support for tracking all stock level changes
-- Author: System
-- Date: 2025-10-18

-- Create stock_adjustments table
CREATE TABLE IF NOT EXISTS stock_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    warehouse_id UUID REFERENCES warehouses(id) ON DELETE SET NULL,
    adjustment_type VARCHAR(50) NOT NULL CHECK (adjustment_type IN ('increase', 'decrease', 'reservation', 'release', 'transfer_in', 'transfer_out', 'correction', 'damage', 'return')),
    quantity INTEGER NOT NULL,
    previous_stock INTEGER NOT NULL,
    new_stock INTEGER NOT NULL,
    reason TEXT,
    reference_type VARCHAR(50), -- 'order', 'reservation', 'transfer', 'manual', etc.
    reference_id UUID, -- ID of the related entity (order_id, reservation_id, etc.)
    adjusted_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    adjusted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Ensure quantity matches the difference
    CONSTRAINT check_quantity_matches_difference CHECK (new_stock - previous_stock = quantity OR previous_stock - new_stock = ABS(quantity))
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_stock_adjustments_tenant_id ON stock_adjustments(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_stock_adjustments_product_id ON stock_adjustments(product_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_stock_adjustments_warehouse_id ON stock_adjustments(warehouse_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_stock_adjustments_adjustment_type ON stock_adjustments(adjustment_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_stock_adjustments_adjusted_by ON stock_adjustments(adjusted_by) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_stock_adjustments_adjusted_at ON stock_adjustments(adjusted_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_stock_adjustments_reference ON stock_adjustments(reference_type, reference_id) WHERE deleted_at IS NULL;

-- Create composite index for common queries
CREATE INDEX IF NOT EXISTS idx_stock_adjustments_tenant_product ON stock_adjustments(tenant_id, product_id, adjusted_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_stock_adjustments_tenant_warehouse_product ON stock_adjustments(tenant_id, warehouse_id, product_id, adjusted_at DESC) WHERE deleted_at IS NULL;

-- Add comments for documentation
COMMENT ON TABLE stock_adjustments IS 'Complete history of all stock level changes for audit and tracking';
COMMENT ON COLUMN stock_adjustments.adjustment_type IS 'Type of adjustment: increase, decrease, reservation, release, transfer_in, transfer_out, correction, damage, return';
COMMENT ON COLUMN stock_adjustments.quantity IS 'Quantity changed (positive for increase, negative for decrease)';
COMMENT ON COLUMN stock_adjustments.previous_stock IS 'Stock level before the adjustment';
COMMENT ON COLUMN stock_adjustments.new_stock IS 'Stock level after the adjustment';
COMMENT ON COLUMN stock_adjustments.reference_type IS 'Type of entity that triggered this adjustment';
COMMENT ON COLUMN stock_adjustments.reference_id IS 'ID of the entity that triggered this adjustment';
