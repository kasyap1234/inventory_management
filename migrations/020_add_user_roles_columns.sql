-- Fix missing columns in user_roles, roles, and users tables for proper RBAC and 2FA support
-- This migration fixes issues discovered during email registration testing

-- 1. Add missing columns to user_roles table for proper role assignment tracking
ALTER TABLE user_roles 
ADD COLUMN IF NOT EXISTS tenant_id UUID,
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS assigned_at TIMESTAMP DEFAULT NOW();

-- 2. Add missing is_active column to roles table
ALTER TABLE roles ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- 3. Add missing 2FA columns to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS two_factor_secret TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS two_factor_enabled BOOLEAN DEFAULT false;

-- Add foreign key constraint for tenant_id if it doesn't exist
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'user_roles_tenant_id_fkey'
        AND table_name = 'user_roles'
    ) THEN
        ALTER TABLE user_roles 
        ADD CONSTRAINT user_roles_tenant_id_fkey 
        FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Update existing user_roles to set tenant_id from users table
UPDATE user_roles ur
SET tenant_id = u.tenant_id
FROM users u
WHERE ur.user_id = u.id
AND ur.tenant_id IS NULL;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_roles_tenant_user ON user_roles (tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_is_active ON user_roles (is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_roles_is_active ON roles (is_active) WHERE is_active = true;

-- Clean up invalid data and make tenant_id required
DELETE FROM user_roles WHERE tenant_id IS NULL;
ALTER TABLE user_roles ALTER COLUMN tenant_id SET NOT NULL;

-- Update the unique constraint to include tenant_id
DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'user_roles_user_id_role_id_key'
        AND table_name = 'user_roles'
    ) THEN
        ALTER TABLE user_roles DROP CONSTRAINT user_roles_user_id_role_id_key;
    END IF;
END $$;

-- Add the new unique constraint if it doesn't exist
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'user_roles_user_id_tenant_id_role_id_key'
        AND table_name = 'user_roles'
    ) THEN
        ALTER TABLE user_roles ADD CONSTRAINT user_roles_user_id_tenant_id_role_id_key UNIQUE (user_id, tenant_id, role_id);
    END IF;
END $$;

-- Add comments
COMMENT ON COLUMN user_roles.tenant_id IS 'Tenant ID for multi-tenancy support';
COMMENT ON COLUMN user_roles.is_active IS 'Whether this role assignment is currently active';
COMMENT ON COLUMN user_roles.assigned_at IS 'When this role was assigned to the user';
COMMENT ON COLUMN roles.is_active IS 'Whether this role is currently active and can be assigned';
COMMENT ON COLUMN users.two_factor_secret IS 'Secret key for TOTP-based two-factor authentication';
COMMENT ON COLUMN users.two_factor_enabled IS 'Whether two-factor authentication is enabled for this user';
