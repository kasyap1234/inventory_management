-- Migration: Default Role and Invitation Roles
-- Adds default_role_id to tenants and role_ids to invitations for automatic role assignment

-- ============================================================================
-- PART 1: Add Default Role to Tenants
-- ============================================================================

-- Add default_role_id column to tenants (role to assign to new users on approval)
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS default_role_id UUID REFERENCES roles(id) ON DELETE SET NULL;

-- Create index for efficient lookups
CREATE INDEX IF NOT EXISTS idx_tenants_default_role ON tenants(default_role_id) WHERE default_role_id IS NOT NULL;

-- ============================================================================
-- PART 2: Add Role IDs to Invitations
-- ============================================================================

-- Add role_ids column to invitations (roles to assign when invitation is accepted)
-- Using JSONB array to store multiple role IDs
ALTER TABLE invitations ADD COLUMN IF NOT EXISTS role_ids JSONB DEFAULT '[]'::jsonb;

-- Create index for efficient queries on role_ids
CREATE INDEX IF NOT EXISTS idx_invitations_role_ids ON invitations USING GIN(role_ids);

-- ============================================================================
-- PART 3: Create Default "member" Role Template
-- ============================================================================

-- For B2B inventory management, the default role should have:
-- - Read access to products, inventory, orders
-- - Create/update for orders (allows placing orders)
-- - Read access to basic info like warehouses, categories

-- This will be applied by the application during tenant onboarding
-- The role priority for "member" is 200 (lower than user=300)

-- Insert default member role permissions into the permission table if they don't exist
INSERT INTO permissions (name, resource, action, description) VALUES
('member.read', 'member', 'read', 'Basic member read access'),
('member.dashboard', 'member', 'dashboard', 'Access to member dashboard')
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- PART 4: Set Default Role for Existing Tenants (if user role exists)
-- ============================================================================

-- For existing tenants, set the "user" role as default if it exists
UPDATE tenants t
SET default_role_id = r.id
FROM roles r
WHERE r.tenant_id = t.id
  AND r.name = 'user'
  AND t.default_role_id IS NULL;

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================
