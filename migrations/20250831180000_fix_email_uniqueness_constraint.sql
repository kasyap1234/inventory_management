-- Fix email uniqueness constraint to be tenant-scoped instead of global
-- This allows users to have the same email in different tenants

-- First, drop the global unique constraint on email (if it exists)
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

-- Drop any global index on email
DROP INDEX IF EXISTS idx_users_email;

-- Add tenant-scoped unique constraint on (tenant_id, email)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'users_tenant_email_unique'
          AND conrelid = 'users'::regclass
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_tenant_email_unique UNIQUE (tenant_id, email);
    END IF;
END
$$;

-- Ensure the composite index exists for performance
CREATE INDEX IF NOT EXISTS idx_users_tenant_email ON users (tenant_id, email);

-- Add explanatory comment
COMMENT ON CONSTRAINT users_tenant_email_unique ON users IS 'Email must be unique within each tenant but can be shared across tenants';