-- Migration: Add tenant_id to batches table for multi-tenant isolation
-- This fixes a critical security vulnerability where batches could be accessed across tenants

-- 1. Add tenant_id column (nullable initially for data migration)
ALTER TABLE batches ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- 2. Populate tenant_id from the related product's tenant_id
UPDATE batches
SET tenant_id = p.tenant_id
FROM products p
WHERE batches.product_id = p.id
AND batches.tenant_id IS NULL;

-- 3. Make tenant_id NOT NULL now that data is migrated (only if there are no null values)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM batches WHERE tenant_id IS NULL) THEN
        ALTER TABLE batches ALTER COLUMN tenant_id SET NOT NULL;
    END IF;
END $$;

-- 4. Add foreign key constraint (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'batches_tenant_id_fkey'
    ) THEN
        ALTER TABLE batches
        ADD CONSTRAINT batches_tenant_id_fkey
        FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- 5. Update the unique constraint to include tenant_id
ALTER TABLE batches DROP CONSTRAINT IF EXISTS batches_product_id_batch_number_key;
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'batches_product_id_batch_number_tenant_key'
    ) THEN
        ALTER TABLE batches ADD CONSTRAINT batches_product_id_batch_number_tenant_key
            UNIQUE(product_id, batch_number, tenant_id);
    END IF;
END $$;

-- 6. Add index on tenant_id for query performance
CREATE INDEX IF NOT EXISTS idx_batches_tenant_id ON batches(tenant_id);

-- 7. Add composite index for common queries
CREATE INDEX IF NOT EXISTS idx_batches_tenant_product ON batches(tenant_id, product_id);
