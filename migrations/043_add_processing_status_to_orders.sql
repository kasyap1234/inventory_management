-- Add 'processing' status to orders table status check constraint
-- The order service uses 'processing' status but it was missing from the original constraint

DO $$
BEGIN
    -- Drop the old constraint and add a new one that includes 'processing'
    ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
    ALTER TABLE orders ADD CONSTRAINT orders_status_check 
        CHECK (status IN ('pending', 'approved', 'processing', 'received', 'shipped', 'delivered', 'cancelled'));
END $$;
