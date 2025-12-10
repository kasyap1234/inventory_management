-- Multi-Tenant B2B RBAC Enhancements
-- Adds platform admin flag and system role protection

-- ============================================================================
-- PART 1: Add Platform Admin Flag to Users
-- ============================================================================

-- Add is_platform_admin column (super admin can manage all tenants)
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_platform_admin BOOLEAN DEFAULT FALSE;

-- Create index for efficient platform admin lookups
CREATE INDEX IF NOT EXISTS idx_users_platform_admin ON users(is_platform_admin) WHERE is_platform_admin = true;

-- ============================================================================
-- PART 2: Add System Role Flag to Roles
-- ============================================================================

-- Add is_system_role column (cannot be deleted by tenant admins)
ALTER TABLE roles ADD COLUMN IF NOT EXISTS is_system_role BOOLEAN DEFAULT FALSE;

-- Mark admin roles as system roles (cannot be deleted)
UPDATE roles SET is_system_role = true WHERE name IN ('admin', 'super_admin');

-- ============================================================================
-- PART 3: Set Default Role Priorities
-- ============================================================================

-- Ensure all roles have appropriate priority values
-- Higher priority = more access, can only assign roles with lower priority
UPDATE roles SET priority = 900 WHERE name = 'admin' AND (priority IS NULL OR priority = 0);
UPDATE roles SET priority = 700 WHERE name = 'manager' AND (priority IS NULL OR priority = 0);
UPDATE roles SET priority = 500 WHERE name IN ('inventory_manager', 'sales') AND (priority IS NULL OR priority = 0);
UPDATE roles SET priority = 400 WHERE name = 'analyst' AND (priority IS NULL OR priority = 0);
UPDATE roles SET priority = 300 WHERE name = 'user' AND (priority IS NULL OR priority = 0);
UPDATE roles SET priority = 100 WHERE name = 'viewer' AND (priority IS NULL OR priority = 0);

-- ============================================================================
-- PART 4: Mark Existing Super Admin (if exists via env var)
-- ============================================================================

-- This will be handled by the application on startup via SeedSuperAdmin function
-- The super admin email comes from SUPER_ADMIN_EMAIL environment variable

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================
