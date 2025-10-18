-- Migration: Create inventory reservations table
-- Description: Adds support for inventory reservation tracking
-- Author: System
-- Date: 2025-10-18

-- Create inventory_reservations table
CREATE TABLE IF NOT EXISTS inventory_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    warehouse_id UUID REFERENCES warehouses(id) ON DELETE SET NULL,
    reservation_id VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    reserved_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reserved_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'expired', 'released', 'committed')),
    order_id UUID REFERENCES orders(id) ON DELETE SET NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Composite unique constraint to prevent duplicate reservations
    CONSTRAINT unique_reservation_id_per_tenant UNIQUE (tenant_id, reservation_id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_inventory_reservations_tenant_id ON inventory_reservations(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_inventory_reservations_product_id ON inventory_reservations(product_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_inventory_reservations_warehouse_id ON inventory_reservations(warehouse_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_inventory_reservations_status ON inventory_reservations(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_inventory_reservations_reserved_by ON inventory_reservations(reserved_by) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_inventory_reservations_order_id ON inventory_reservations(order_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_inventory_reservations_expires_at ON inventory_reservations(expires_at) WHERE deleted_at IS NULL AND status = 'active';

-- Create composite index for common queries
CREATE INDEX IF NOT EXISTS idx_inventory_reservations_tenant_product ON inventory_reservations(tenant_id, product_id, status) WHERE deleted_at IS NULL;

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_inventory_reservations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_inventory_reservations_updated_at
    BEFORE UPDATE ON inventory_reservations
    FOR EACH ROW
    EXECUTE FUNCTION update_inventory_reservations_updated_at();

-- Add comments for documentation
COMMENT ON TABLE inventory_reservations IS 'Tracks inventory reservations for orders and other purposes';
COMMENT ON COLUMN inventory_reservations.reservation_id IS 'Unique identifier for the reservation (can be order number, etc.)';
COMMENT ON COLUMN inventory_reservations.status IS 'Reservation status: active, expired, released, or committed';
COMMENT ON COLUMN inventory_reservations.expires_at IS 'Optional expiration time for the reservation';
COMMENT ON COLUMN inventory_reservations.order_id IS 'Optional reference to the order that created this reservation';
