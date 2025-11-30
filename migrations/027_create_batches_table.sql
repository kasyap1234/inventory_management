-- Migration: Create batches table and update products for agro-tech features

-- 1. Create batches table
CREATE TABLE IF NOT EXISTS batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    batch_number VARCHAR(100) NOT NULL,
    quantity INTEGER DEFAULT 0 CHECK (quantity >= 0),
    expiry_date DATE,
    manufacturing_date DATE,
    location VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'expired', 'quarantined', 'recalled')),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(product_id, batch_number)
);

CREATE INDEX IF NOT EXISTS idx_batches_product_id ON batches(product_id);
CREATE INDEX IF NOT EXISTS idx_batches_expiry_date ON batches(expiry_date);
CREATE INDEX IF NOT EXISTS idx_batches_batch_number ON batches(batch_number);

-- 2. Add hazardous material fields to products
ALTER TABLE products 
ADD COLUMN IF NOT EXISTS is_hazardous BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS hazard_class VARCHAR(50),
ADD COLUMN IF NOT EXISTS sds_url TEXT,
ADD COLUMN IF NOT EXISTS active_ingredients TEXT;

-- 3. Migrate existing data (if any) from products to batches
-- We create a default batch for existing products that have quantity > 0
INSERT INTO batches (product_id, batch_number, quantity, expiry_date, status)
SELECT 
    id, 
    COALESCE(batch_number, 'LEGACY-' || SUBSTRING(id::text, 1, 8)), 
    quantity, 
    expiry_date,
    'active'
FROM products 
WHERE quantity > 0;

-- 4. Make old columns nullable (or drop them later, for now just nullable to avoid breaking immediately)
ALTER TABLE products ALTER COLUMN batch_number DROP NOT NULL;
ALTER TABLE products ALTER COLUMN expiry_date DROP NOT NULL;
-- We keep quantity in products as a cached total for performance, but it should be updated via triggers or application logic.
-- For now, we trust the application to maintain it.
