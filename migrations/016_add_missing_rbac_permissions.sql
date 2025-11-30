-- Add Missing RBAC Permissions
-- This migration adds comprehensive permissions for distributor management, warehouse management, 
-- analytics access, and subscription management

-- First, update the existing permissions table structure to include resource, action, and conditions
ALTER TABLE permissions
ADD COLUMN IF NOT EXISTS resource VARCHAR(50),
ADD COLUMN IF NOT EXISTS action VARCHAR(50),
ADD COLUMN IF NOT EXISTS conditions JSONB DEFAULT '{}',
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();

-- Update existing permissions to have resource and action based on name pattern
UPDATE permissions SET
    resource = CASE
        WHEN name LIKE 'read_%' THEN 'general'
        WHEN name LIKE 'create_%' THEN 'general'
        WHEN name LIKE 'update_%' THEN 'general'
        WHEN name LIKE 'delete_%' THEN 'general'
        WHEN name LIKE 'manage_%' THEN 'general'
        WHEN name LIKE 'users:%' THEN 'user'
        WHEN name LIKE 'tenants:%' THEN 'tenant'
        ELSE 'general'
    END,
    action = CASE
        WHEN name LIKE 'tenants:%' THEN regexp_replace(name, 'tenants:', '')
        WHEN name LIKE 'users:%' THEN regexp_replace(name, 'users:', '')
        ELSE regexp_replace(name, '_.*', '')
    END
WHERE resource IS NULL OR action IS NULL;

-- Create role_permissions junction table if it doesn't exist
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    conditions JSONB DEFAULT '{}',
    granted_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

-- Add missing columns to existing role_permissions table
ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS conditions JSONB DEFAULT '{}';
ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS granted_at TIMESTAMP DEFAULT NOW();

-- Insert Distributor Management Permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('distributor.create', 'distributor', 'create', 'Create new distributors'),
('distributor.read', 'distributor', 'read', 'View distributor information'),
('distributor.update', 'distributor', 'update', 'Update distributor information'),
('distributor.delete', 'distributor', 'delete', 'Delete distributors'),
('distributor.list', 'distributor', 'list', 'List all distributors'),
('distributor.manage_contracts', 'distributor', 'manage_contracts', 'Manage distributor contracts'),
('distributor.view_performance', 'distributor', 'view_performance', 'View distributor performance metrics'),
('distributor.manage_payments', 'distributor', 'manage_payments', 'Manage distributor payments')
ON CONFLICT (name) DO NOTHING;

-- Insert Warehouse Management Permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('warehouse.create', 'warehouse', 'create', 'Create new warehouses'),
('warehouse.read', 'warehouse', 'read', 'View warehouse information'),
('warehouse.update', 'warehouse', 'update', 'Update warehouse information'),
('warehouse.delete', 'warehouse', 'delete', 'Delete warehouses'),
('warehouse.list', 'warehouse', 'list', 'List all warehouses'),
('warehouse.manage_inventory', 'warehouse', 'manage_inventory', 'Manage warehouse inventory'),
('warehouse.view_capacity', 'warehouse', 'view_capacity', 'View warehouse capacity and utilization'),
('warehouse.manage_transfers', 'warehouse', 'manage_transfers', 'Manage inventory transfers between warehouses'),
('warehouse.view_reports', 'warehouse', 'view_reports', 'View warehouse reports and analytics')
ON CONFLICT (name) DO NOTHING;

-- Insert Analytics Access Permissions with Data-Level Restrictions
INSERT INTO permissions (name, resource, action, conditions, description) VALUES
('analytics.view_sales', 'analytics', 'view_sales', '{"data_scope": "own"}', 'View sales analytics for own data'),
('analytics.view_sales_all', 'analytics', 'view_sales', '{"data_scope": "all"}', 'View sales analytics for all data'),
('analytics.view_inventory', 'analytics', 'view_inventory', '{"data_scope": "own"}', 'View inventory analytics for own data'),
('analytics.view_inventory_all', 'analytics', 'view_inventory', '{"data_scope": "all"}', 'View inventory analytics for all data'),
('analytics.view_financial', 'analytics', 'view_financial', '{"data_scope": "own"}', 'View financial analytics for own data'),
('analytics.view_financial_all', 'analytics', 'view_financial', '{"data_scope": "all"}', 'View financial analytics for all data'),
('analytics.export_data', 'analytics', 'export', '{"formats": ["csv", "excel"]}', 'Export analytics data'),
('analytics.create_reports', 'analytics', 'create_reports', '{}', 'Create custom analytics reports'),
('analytics.manage_dashboards', 'analytics', 'manage_dashboards', '{}', 'Create and manage analytics dashboards')
ON CONFLICT (name) DO NOTHING;

-- Insert Subscription Management Permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('subscription.view', 'subscription', 'view', 'View subscription information'),
('subscription.manage', 'subscription', 'manage', 'Manage subscription plans and billing'),
('subscription.upgrade', 'subscription', 'upgrade', 'Upgrade subscription plans'),
('subscription.downgrade', 'subscription', 'downgrade', 'Downgrade subscription plans'),
('subscription.cancel', 'subscription', 'cancel', 'Cancel subscriptions'),
('subscription.view_billing', 'subscription', 'view_billing', 'View billing history and invoices'),
('subscription.manage_payment_methods', 'subscription', 'manage_payment_methods', 'Manage payment methods'),
('subscription.view_usage', 'subscription', 'view_usage', 'View subscription usage metrics')
ON CONFLICT (name) DO NOTHING;

-- Insert Enhanced Product Management Permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('product.manage_images', 'product', 'manage_images', 'Upload and manage product images'),
('product.bulk_import', 'product', 'bulk_import', 'Import products in bulk'),
('product.bulk_export', 'product', 'bulk_export', 'Export products in bulk'),
('product.manage_categories', 'product', 'manage_categories', 'Manage product categories'),
('product.view_analytics', 'product', 'view_analytics', 'View product performance analytics')
ON CONFLICT (name) DO NOTHING;

-- Insert Order Management Permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('order.approve', 'order', 'approve', 'Approve pending orders'),
('order.reject', 'order', 'reject', 'Reject orders'),
('order.fulfill', 'order', 'fulfill', 'Mark orders as fulfilled'),
('order.track', 'order', 'track', 'Track order status and shipments'),
('order.manage_returns', 'order', 'manage_returns', 'Manage order returns and refunds')
ON CONFLICT (name) DO NOTHING;

-- Insert Notification System Permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('notification.create_templates', 'notification', 'create_templates', 'Create notification templates'),
('notification.manage_webhooks', 'notification', 'manage_webhooks', 'Manage webhook subscriptions'),
('notification.configure_alerts', 'notification', 'configure_alerts', 'Configure alert rules'),
('notification.view_delivery_logs', 'notification', 'view_delivery_logs', 'View notification delivery logs'),
('notification.send_manual', 'notification', 'send_manual', 'Send manual notifications')
ON CONFLICT (name) DO NOTHING;

-- Insert User Management Permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('user.invite', 'user', 'invite', 'Invite new users to the system'),
('user.deactivate', 'user', 'deactivate', 'Deactivate user accounts'),
('user.reset_password', 'user', 'reset_password', 'Reset user passwords'),
('user.manage_roles', 'user', 'manage_roles', 'Assign and manage user roles'),
('user.view_activity', 'user', 'view_activity', 'View user activity logs')
ON CONFLICT (name) DO NOTHING;

-- Create default role assignments
-- First, ensure we have basic roles (using the default tenant ID from complete_auth_schema.sql)
INSERT INTO roles (id, tenant_id, name, description) VALUES
(gen_random_uuid(), 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'super_admin', 'Super Administrator with all permissions'),
(gen_random_uuid(), 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'admin', 'Administrator with most permissions'),
(gen_random_uuid(), 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'manager', 'Manager with operational permissions'),
(gen_random_uuid(), 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'operator', 'Operator with basic operational permissions'),
(gen_random_uuid(), 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'viewer', 'Read-only access to most resources')
ON CONFLICT (tenant_id, name) DO NOTHING;

-- Ensure UNIQUE constraint exists on role_permissions
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT constraint_name FROM information_schema.table_constraints
        WHERE table_name = 'role_permissions' AND constraint_type = 'UNIQUE'
    ) THEN
        ALTER TABLE role_permissions ADD CONSTRAINT role_permissions_unique UNIQUE (role_id, permission_id);
    END IF;
END $$;

-- Assign permissions to Super Admin role (all permissions)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'super_admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign permissions to Admin role (most permissions except super admin specific ones)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin'
  AND p.name NOT IN ('user.delete', 'subscription.cancel')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign permissions to Manager role (operational permissions)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'manager'
  AND p.resource IN ('product', 'order', 'warehouse', 'distributor')
  AND p.action IN ('create', 'read', 'update', 'list', 'manage_inventory', 'approve', 'fulfill')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign permissions to Operator role (basic operational permissions)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'operator'
  AND p.resource IN ('product', 'order', 'warehouse')
  AND p.action IN ('read', 'list', 'update', 'manage_inventory')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign permissions to Viewer role (read-only permissions)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'viewer'
  AND p.action IN ('read', 'list', 'view')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions (resource, action);
CREATE INDEX IF NOT EXISTS idx_permissions_name ON permissions (name);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions (role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions (permission_id);

-- Create a view for easy permission checking
CREATE OR REPLACE VIEW user_permissions AS
SELECT 
    ur.user_id,
    u.tenant_id,
    p.name as permission_name,
    p.resource,
    p.action,
    p.conditions as permission_conditions,
    rp.conditions as role_conditions,
    r.name as role_name
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
JOIN roles r ON ur.role_id = r.id
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id;

-- Drop any existing versions of user_has_permission to avoid conflicts
DROP FUNCTION IF EXISTS user_has_permission(uuid, uuid, varchar, jsonb);
DROP FUNCTION IF EXISTS user_has_permission(uuid, uuid, varchar);

-- Create a function to check if a user has a specific permission
CREATE OR REPLACE FUNCTION user_has_permission(
    p_user_id UUID,
    p_tenant_id UUID,
    p_permission_name VARCHAR,
    p_context JSONB DEFAULT '{}'
)
RETURNS BOOLEAN AS $$
DECLARE
    permission_exists BOOLEAN := FALSE;
BEGIN
    SELECT EXISTS(
        SELECT 1 
        FROM user_permissions up
        WHERE up.user_id = p_user_id 
          AND up.tenant_id = p_tenant_id 
          AND up.permission_name = p_permission_name
          -- Add context-aware permission checking here if needed
    ) INTO permission_exists;
    
    RETURN permission_exists;
END;
$$ LANGUAGE plpgsql;

-- Create a function to get all permissions for a user
CREATE OR REPLACE FUNCTION get_user_permissions(
    p_user_id UUID,
    p_tenant_id UUID
)
RETURNS TABLE (
    permission_name VARCHAR,
    resource VARCHAR,
    action VARCHAR,
    conditions JSONB,
    role_name VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        up.permission_name,
        up.resource,
        up.action,
        COALESCE(up.role_conditions, up.permission_conditions) as conditions,
        up.role_name
    FROM user_permissions up
    WHERE up.user_id = p_user_id 
      AND up.tenant_id = p_tenant_id
    ORDER BY up.resource, up.action;
END;
$$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON TABLE permissions IS 'System permissions that can be assigned to roles';
COMMENT ON TABLE role_permissions IS 'Junction table linking roles to permissions';
COMMENT ON VIEW user_permissions IS 'View showing all permissions for users through their roles';
COMMENT ON FUNCTION user_has_permission IS 'Check if a user has a specific permission';
COMMENT ON FUNCTION get_user_permissions IS 'Get all permissions for a user';
