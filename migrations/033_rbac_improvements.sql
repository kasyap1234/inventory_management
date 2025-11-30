-- RBAC Improvements Migration
-- This migration standardizes permission naming, adds role hierarchy support,
-- and optimizes permission queries

-- ============================================================================
-- PART 1: Standardize Permission Naming to Dot Notation
-- ============================================================================

-- Convert all colon-notation permissions to dot-notation
-- e.g., "users:list" -> "user.list", "products:create" -> "product.create"

-- First, delete any colon-notation permissions that already have a dot-notation equivalent
-- This prevents unique constraint violations during the UPDATE
DELETE FROM permissions WHERE name LIKE '%:%' 
AND REPLACE(
    CASE 
        WHEN name LIKE 'users:%' THEN regexp_replace(name, '^users:', 'user.')
        WHEN name LIKE 'products:%' THEN regexp_replace(name, '^products:', 'product.')
        WHEN name LIKE 'orders:%' THEN regexp_replace(name, '^orders:', 'order.')
        WHEN name LIKE 'invoices:%' THEN regexp_replace(name, '^invoices:', 'invoice.')
        WHEN name LIKE 'inventories:%' THEN regexp_replace(name, '^inventories:', 'inventory.')
        WHEN name LIKE 'warehouses:%' THEN regexp_replace(name, '^warehouses:', 'warehouse.')
        WHEN name LIKE 'distributors:%' THEN regexp_replace(name, '^distributors:', 'distributor.')
        WHEN name LIKE 'suppliers:%' THEN regexp_replace(name, '^suppliers:', 'supplier.')
        WHEN name LIKE 'categories:%' THEN regexp_replace(name, '^categories:', 'category.')
        WHEN name LIKE 'tenants:%' THEN regexp_replace(name, '^tenants:', 'tenant.')
        ELSE REPLACE(name, ':', '.')
    END, ':', '.')
IN (SELECT name FROM permissions WHERE name NOT LIKE '%:%');

-- Now update remaining colon-notation permissions to dot notation (also convert plural to singular)
UPDATE permissions SET 
    name = CASE 
        WHEN name LIKE 'users:%' THEN regexp_replace(name, '^users:', 'user.')
        WHEN name LIKE 'products:%' THEN regexp_replace(name, '^products:', 'product.')
        WHEN name LIKE 'orders:%' THEN regexp_replace(name, '^orders:', 'order.')
        WHEN name LIKE 'invoices:%' THEN regexp_replace(name, '^invoices:', 'invoice.')
        WHEN name LIKE 'inventories:%' THEN regexp_replace(name, '^inventories:', 'inventory.')
        WHEN name LIKE 'inventory:%' THEN regexp_replace(name, '^inventory:', 'inventory.')
        WHEN name LIKE 'warehouses:%' THEN regexp_replace(name, '^warehouses:', 'warehouse.')
        WHEN name LIKE 'distributors:%' THEN regexp_replace(name, '^distributors:', 'distributor.')
        WHEN name LIKE 'suppliers:%' THEN regexp_replace(name, '^suppliers:', 'supplier.')
        WHEN name LIKE 'categories:%' THEN regexp_replace(name, '^categories:', 'category.')
        WHEN name LIKE 'tenants:%' THEN regexp_replace(name, '^tenants:', 'tenant.')
        WHEN name LIKE 'analytics:%' THEN regexp_replace(name, '^analytics:', 'analytics.')
        ELSE REPLACE(name, ':', '.')
    END,
    resource = CASE 
        WHEN name LIKE 'users:%' THEN 'user'
        WHEN name LIKE 'products:%' THEN 'product'
        WHEN name LIKE 'orders:%' THEN 'order'
        WHEN name LIKE 'invoices:%' THEN 'invoice'
        WHEN name LIKE 'inventories:%' OR name LIKE 'inventory:%' THEN 'inventory'
        WHEN name LIKE 'warehouses:%' THEN 'warehouse'
        WHEN name LIKE 'distributors:%' THEN 'distributor'
        WHEN name LIKE 'suppliers:%' THEN 'supplier'
        WHEN name LIKE 'categories:%' THEN 'category'
        WHEN name LIKE 'tenants:%' THEN 'tenant'
        WHEN name LIKE 'analytics:%' THEN 'analytics'
        ELSE resource
    END,
    action = regexp_replace(name, '^[^:]+:', '')
WHERE name LIKE '%:%';

-- First, delete any plural-form permissions that already have singular equivalents
-- This prevents unique constraint violations during the UPDATE below
DELETE FROM permissions p1
WHERE p1.name LIKE 'users.%' OR p1.name LIKE 'products.%' OR p1.name LIKE 'orders.%' 
   OR p1.name LIKE 'invoices.%' OR p1.name LIKE 'inventories.%' OR p1.name LIKE 'warehouses.%'
   OR p1.name LIKE 'distributors.%' OR p1.name LIKE 'suppliers.%' OR p1.name LIKE 'categories.%' 
   OR p1.name LIKE 'tenants.%'
AND EXISTS (
    SELECT 1 FROM permissions p2 WHERE p2.name = 
        CASE
            WHEN p1.name LIKE 'users.%' THEN regexp_replace(p1.name, '^users\.', 'user.')
            WHEN p1.name LIKE 'products.%' THEN regexp_replace(p1.name, '^products\.', 'product.')
            WHEN p1.name LIKE 'orders.%' THEN regexp_replace(p1.name, '^orders\.', 'order.')
            WHEN p1.name LIKE 'invoices.%' THEN regexp_replace(p1.name, '^invoices\.', 'invoice.')
            WHEN p1.name LIKE 'inventories.%' THEN regexp_replace(p1.name, '^inventories\.', 'inventory.')
            WHEN p1.name LIKE 'warehouses.%' THEN regexp_replace(p1.name, '^warehouses\.', 'warehouse.')
            WHEN p1.name LIKE 'distributors.%' THEN regexp_replace(p1.name, '^distributors\.', 'distributor.')
            WHEN p1.name LIKE 'suppliers.%' THEN regexp_replace(p1.name, '^suppliers\.', 'supplier.')
            WHEN p1.name LIKE 'categories.%' THEN regexp_replace(p1.name, '^categories\.', 'category.')
            WHEN p1.name LIKE 'tenants.%' THEN regexp_replace(p1.name, '^tenants\.', 'tenant.')
            ELSE p1.name
        END
);

-- Also convert plural resource names to singular in the name field
UPDATE permissions SET 
    name = CASE
        WHEN name LIKE 'users.%' THEN regexp_replace(name, '^users\.', 'user.')
        WHEN name LIKE 'products.%' THEN regexp_replace(name, '^products\.', 'product.')
        WHEN name LIKE 'orders.%' THEN regexp_replace(name, '^orders\.', 'order.')
        WHEN name LIKE 'invoices.%' THEN regexp_replace(name, '^invoices\.', 'invoice.')
        WHEN name LIKE 'inventories.%' THEN regexp_replace(name, '^inventories\.', 'inventory.')
        WHEN name LIKE 'warehouses.%' THEN regexp_replace(name, '^warehouses\.', 'warehouse.')
        WHEN name LIKE 'distributors.%' THEN regexp_replace(name, '^distributors\.', 'distributor.')
        WHEN name LIKE 'suppliers.%' THEN regexp_replace(name, '^suppliers\.', 'supplier.')
        WHEN name LIKE 'categories.%' THEN regexp_replace(name, '^categories\.', 'category.')
        WHEN name LIKE 'tenants.%' THEN regexp_replace(name, '^tenants\.', 'tenant.')
        ELSE name
    END,
    resource = CASE 
        WHEN resource = 'users' THEN 'user'
        WHEN resource = 'products' THEN 'product'
        WHEN resource = 'orders' THEN 'order'
        WHEN resource = 'invoices' THEN 'invoice'
        WHEN resource = 'inventories' THEN 'inventory'
        WHEN resource = 'warehouses' THEN 'warehouse'
        WHEN resource = 'distributors' THEN 'distributor'
        WHEN resource = 'suppliers' THEN 'supplier'
        WHEN resource = 'categories' THEN 'category'
        WHEN resource = 'tenants' THEN 'tenant'
        ELSE resource
    END
WHERE resource IN ('users', 'products', 'orders', 'invoices', 'inventories', 
                   'warehouses', 'distributors', 'suppliers', 'categories', 'tenants');

-- Delete legacy underscore-prefix permissions if dot-notation equivalents already exist
DELETE FROM permissions WHERE name IN ('read_users', 'create_users', 'update_users', 'delete_users',
                                        'read_products', 'create_products', 'update_products', 'delete_products', 'manage_products')
AND (
    (name = 'read_users' AND EXISTS (SELECT 1 FROM permissions WHERE name = 'user.read')) OR
    (name = 'create_users' AND EXISTS (SELECT 1 FROM permissions WHERE name = 'user.create')) OR
    (name = 'update_users' AND EXISTS (SELECT 1 FROM permissions WHERE name = 'user.update')) OR
    (name = 'delete_users' AND EXISTS (SELECT 1 FROM permissions WHERE name = 'user.delete')) OR
    (name = 'read_products' AND EXISTS (SELECT 1 FROM permissions WHERE name = 'product.read')) OR
    (name = 'create_products' AND EXISTS (SELECT 1 FROM permissions WHERE name = 'product.create')) OR
    (name = 'update_products' AND EXISTS (SELECT 1 FROM permissions WHERE name = 'product.update')) OR
    (name = 'delete_products' AND EXISTS (SELECT 1 FROM permissions WHERE name = 'product.delete')) OR
    (name = 'manage_products' AND EXISTS (SELECT 1 FROM permissions WHERE name = 'product.manage'))
);

-- Convert legacy underscore-prefix permissions (read_users -> user.read)
UPDATE permissions SET
    name = CASE
        WHEN name = 'read_users' THEN 'user.read'
        WHEN name = 'create_users' THEN 'user.create'
        WHEN name = 'update_users' THEN 'user.update'
        WHEN name = 'delete_users' THEN 'user.delete'
        WHEN name = 'read_products' THEN 'product.read'
        WHEN name = 'create_products' THEN 'product.create'
        WHEN name = 'update_products' THEN 'product.update'
        WHEN name = 'delete_products' THEN 'product.delete'
        WHEN name = 'manage_products' THEN 'product.manage'
        ELSE name
    END,
    resource = CASE 
        WHEN name LIKE 'read_%' OR name LIKE 'create_%' OR name LIKE 'update_%' OR name LIKE 'delete_%' OR name LIKE 'manage_%' THEN 
            regexp_replace(regexp_replace(name, '^(read|create|update|delete|manage)_', ''), 's$', '')
        ELSE resource
    END,
    action = CASE 
        WHEN name LIKE 'read_%' THEN 'read'
        WHEN name LIKE 'create_%' THEN 'create'
        WHEN name LIKE 'update_%' THEN 'update'
        WHEN name LIKE 'delete_%' THEN 'delete'
        WHEN name LIKE 'manage_%' THEN 'manage'
        ELSE action
    END
WHERE name LIKE 'read_%' OR name LIKE 'create_%' OR name LIKE 'update_%' 
   OR name LIKE 'delete_%' OR name LIKE 'manage_%';

-- Remove duplicate permissions (keep the one with the earliest created_at)
DELETE FROM permissions p1
USING permissions p2
WHERE p1.id > p2.id 
  AND p1.name = p2.name;

-- ============================================================================
-- PART 2: Add Role Hierarchy Support
-- ============================================================================

-- Add parent_role_id column for role inheritance
ALTER TABLE roles ADD COLUMN IF NOT EXISTS parent_role_id UUID REFERENCES roles(id) ON DELETE SET NULL;

-- Add index for efficient hierarchy traversal
CREATE INDEX IF NOT EXISTS idx_roles_parent_role_id ON roles(parent_role_id) WHERE parent_role_id IS NOT NULL;

-- Add priority column for role ordering (higher priority takes precedence in permission conflicts)
ALTER TABLE roles ADD COLUMN IF NOT EXISTS priority INTEGER DEFAULT 0;

-- Create index for priority-based role ordering
CREATE INDEX IF NOT EXISTS idx_roles_priority ON roles(tenant_id, priority DESC);

-- ============================================================================
-- PART 3: Create Optimized Permission Lookup Function
-- ============================================================================

-- Drop existing function if it exists (for clean update)
DROP FUNCTION IF EXISTS get_user_all_permissions(uuid, uuid);

-- Create optimized function to get all user permissions in a single query
-- This eliminates the N+1 query problem by joining all tables at once
CREATE OR REPLACE FUNCTION get_user_all_permissions(p_user_id UUID, p_tenant_id UUID)
RETURNS TABLE (
    permission_name VARCHAR(255),
    permission_resource VARCHAR(50),
    permission_action VARCHAR(50),
    permission_conditions JSONB,
    role_name VARCHAR(100),
    role_priority INTEGER
) AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE role_hierarchy AS (
        -- Base case: get user's directly assigned roles
        SELECT r.id, r.name, r.priority, r.parent_role_id, 0 as depth
        FROM roles r
        INNER JOIN user_roles ur ON r.id = ur.role_id
        WHERE ur.user_id = p_user_id 
          AND ur.tenant_id = p_tenant_id
          AND r.tenant_id = p_tenant_id
          AND r.is_active = true
          AND COALESCE(ur.is_active, true) = true
        
        UNION ALL
        
        -- Recursive case: get parent roles
        SELECT r.id, r.name, r.priority, r.parent_role_id, rh.depth + 1
        FROM roles r
        INNER JOIN role_hierarchy rh ON r.id = rh.parent_role_id
        WHERE r.tenant_id = p_tenant_id
          AND r.is_active = true
          AND rh.depth < 10  -- Prevent infinite recursion (max 10 levels)
    )
    SELECT DISTINCT ON (p.name)
        p.name::VARCHAR(255) as permission_name,
        COALESCE(p.resource, 'general')::VARCHAR(50) as permission_resource,
        COALESCE(p.action, 'access')::VARCHAR(50) as permission_action,
        COALESCE(rp.conditions, p.conditions, '{}'::jsonb) as permission_conditions,
        rh.name::VARCHAR(100) as role_name,
        rh.priority as role_priority
    FROM role_hierarchy rh
    INNER JOIN role_permissions rp ON rh.id = rp.role_id
    INNER JOIN permissions p ON rp.permission_id = p.id
    ORDER BY p.name, rh.priority DESC, rh.depth ASC;
END;
$$ LANGUAGE plpgsql STABLE;

-- ============================================================================
-- PART 4: Create Permission Check Function (Optimized with Caching Hint)
-- ============================================================================

DROP FUNCTION IF EXISTS user_has_permission(uuid, uuid, varchar);

CREATE OR REPLACE FUNCTION user_has_permission(
    p_user_id UUID, 
    p_tenant_id UUID, 
    p_permission_name VARCHAR(255)
)
RETURNS BOOLEAN AS $$
DECLARE
    v_has_permission BOOLEAN := false;
    v_permission_resource VARCHAR(50);
    v_permission_action VARCHAR(50);
BEGIN
    -- Parse the permission name to extract resource and action
    v_permission_resource := split_part(p_permission_name, '.', 1);
    v_permission_action := split_part(p_permission_name, '.', 2);
    
    -- Check for exact permission match or wildcard patterns
    SELECT EXISTS (
        SELECT 1
        FROM get_user_all_permissions(p_user_id, p_tenant_id) up
        WHERE 
            -- Exact match
            up.permission_name = p_permission_name
            -- Full wildcard (admin has *)
            OR up.permission_name = '*'
            -- Resource wildcard (e.g., "product.*" matches "product.create")
            OR (up.permission_name = v_permission_resource || '.*')
            -- Action wildcard (e.g., "*.read" matches "product.read")
            OR (up.permission_name = '*.' || v_permission_action)
    ) INTO v_has_permission;
    
    RETURN v_has_permission;
END;
$$ LANGUAGE plpgsql STABLE;

-- ============================================================================
-- PART 5: Permission Validation (Application Layer)
-- ============================================================================
-- NOTE: Permission name validation is handled at the application layer rather than
-- via database triggers. This ensures migrations can be re-run idempotently without
-- conflicts when earlier migrations use legacy permission naming conventions.
-- The RBAC service validates permission names at runtime.

-- Drop any existing validation trigger to ensure migration idempotency
DROP TRIGGER IF EXISTS trigger_validate_permission_name ON permissions;
DROP FUNCTION IF EXISTS validate_permission_name();

-- ============================================================================
-- PART 6: Insert Standard Permissions (Idempotent)
-- ============================================================================

-- Insert all standard permissions with consistent dot notation
INSERT INTO permissions (name, resource, action, description) VALUES
-- User management
('user.list', 'user', 'list', 'List all users'),
('user.create', 'user', 'create', 'Create new users'),
('user.read', 'user', 'read', 'View user details'),
('user.update', 'user', 'update', 'Update user information'),
('user.delete', 'user', 'delete', 'Delete users'),
('user.approve', 'user', 'approve', 'Approve pending users'),
('user.invite', 'user', 'invite', 'Invite new users'),
('user.manage_roles', 'user', 'manage_roles', 'Assign roles to users'),

-- Product management
('product.list', 'product', 'list', 'List all products'),
('product.create', 'product', 'create', 'Create new products'),
('product.read', 'product', 'read', 'View product details'),
('product.update', 'product', 'update', 'Update product information'),
('product.delete', 'product', 'delete', 'Delete products'),
('product.search', 'product', 'search', 'Search products'),
('product.bulk_create', 'product', 'bulk_create', 'Bulk create products'),
('product.bulk_update', 'product', 'bulk_update', 'Bulk update products'),
('product.bulk_price_update', 'product', 'bulk_price_update', 'Bulk update product prices'),

-- Inventory management
('inventory.list', 'inventory', 'list', 'List inventory items'),
('inventory.create', 'inventory', 'create', 'Create inventory records'),
('inventory.read', 'inventory', 'read', 'View inventory details'),
('inventory.update', 'inventory', 'update', 'Update inventory'),
('inventory.delete', 'inventory', 'delete', 'Delete inventory records'),
('inventory.adjust', 'inventory', 'adjust', 'Adjust stock levels'),
('inventory.search', 'inventory', 'search', 'Search inventory'),

-- Order management
('order.list', 'order', 'list', 'List all orders'),
('order.create', 'order', 'create', 'Create new orders'),
('order.read', 'order', 'read', 'View order details'),
('order.update', 'order', 'update', 'Update order information'),
('order.delete', 'order', 'delete', 'Delete orders'),
('order.approve', 'order', 'approve', 'Approve pending orders'),
('order.process', 'order', 'process', 'Process orders'),
('order.receive', 'order', 'receive', 'Receive orders'),
('order.ship', 'order', 'ship', 'Ship orders'),
('order.deliver', 'order', 'deliver', 'Mark orders as delivered'),
('order.cancel', 'order', 'cancel', 'Cancel orders'),

-- Invoice management
('invoice.list', 'invoice', 'list', 'List all invoices'),
('invoice.create', 'invoice', 'create', 'Create new invoices'),
('invoice.read', 'invoice', 'read', 'View invoice details'),
('invoice.update', 'invoice', 'update', 'Update invoice information'),
('invoice.delete', 'invoice', 'delete', 'Delete invoices'),
('invoice.bulk_create', 'invoice', 'bulk_create', 'Bulk create invoices'),
('invoice.update_status', 'invoice', 'update_status', 'Update invoice status'),
('invoice.generate_pdf', 'invoice', 'generate_pdf', 'Generate invoice PDF'),

-- Warehouse management
('warehouse.list', 'warehouse', 'list', 'List warehouses'),
('warehouse.create', 'warehouse', 'create', 'Create warehouses'),
('warehouse.read', 'warehouse', 'read', 'View warehouse details'),
('warehouse.update', 'warehouse', 'update', 'Update warehouses'),
('warehouse.delete', 'warehouse', 'delete', 'Delete warehouses'),

-- Distributor management
('distributor.list', 'distributor', 'list', 'List distributors'),
('distributor.create', 'distributor', 'create', 'Create distributors'),
('distributor.read', 'distributor', 'read', 'View distributor details'),
('distributor.update', 'distributor', 'update', 'Update distributors'),
('distributor.delete', 'distributor', 'delete', 'Delete distributors'),

-- Supplier management
('supplier.list', 'supplier', 'list', 'List suppliers'),
('supplier.create', 'supplier', 'create', 'Create suppliers'),
('supplier.read', 'supplier', 'read', 'View supplier details'),
('supplier.update', 'supplier', 'update', 'Update suppliers'),
('supplier.delete', 'supplier', 'delete', 'Delete suppliers'),

-- Category management
('category.list', 'category', 'list', 'List categories'),
('category.create', 'category', 'create', 'Create categories'),
('category.read', 'category', 'read', 'View category details'),
('category.update', 'category', 'update', 'Update categories'),
('category.delete', 'category', 'delete', 'Delete categories'),

-- Tenant management (admin only)
('tenant.list', 'tenant', 'list', 'List all tenants'),
('tenant.create', 'tenant', 'create', 'Create new tenants'),
('tenant.read', 'tenant', 'read', 'View tenant details'),
('tenant.update', 'tenant', 'update', 'Update tenant information'),
('tenant.delete', 'tenant', 'delete', 'Delete tenants'),
('tenant.manage_settings', 'tenant', 'manage_settings', 'Manage tenant settings'),

-- Analytics (admin/manager)
('analytics.read', 'analytics', 'read', 'View analytics'),
('analytics.dashboard', 'analytics', 'dashboard', 'View dashboard analytics'),
('analytics.export', 'analytics', 'export', 'Export analytics data'),

-- Audit logs
('audit.read', 'audit', 'read', 'View audit logs'),
('audit.export', 'audit', 'export', 'Export audit logs'),

-- System administration
('system.admin', 'system', 'admin', 'Full system administration access'),

-- Batch management
('batch.list', 'batch', 'list', 'List batches'),
('batch.create', 'batch', 'create', 'Create batches'),
('batch.read', 'batch', 'read', 'View batch details'),
('batch.update', 'batch', 'update', 'Update batches'),
('batch.delete', 'batch', 'delete', 'Delete batches'),

-- Notification management
('notification.send', 'notification', 'send', 'Send notifications'),
('notification.manage', 'notification', 'manage', 'Manage notification settings'),

-- Webhook management
('webhook.test', 'webhook', 'test', 'Test webhooks')

ON CONFLICT (name) DO UPDATE SET
    resource = EXCLUDED.resource,
    action = EXCLUDED.action,
    description = EXCLUDED.description,
    updated_at = NOW();

-- ============================================================================
-- PART 7: Create Indexes for Query Optimization
-- ============================================================================

-- Composite index for permission checks (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_user_roles_user_tenant_active 
    ON user_roles(user_id, tenant_id) 
    WHERE COALESCE(is_active, true) = true;

-- Composite index for role permission lookups
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_permission 
    ON role_permissions(role_id, permission_id);

-- Index for permission name lookups
CREATE INDEX IF NOT EXISTS idx_permissions_name ON permissions(name);

-- Index for permission resource lookups (for wildcard matching)
CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions(resource);

-- ============================================================================
-- PART 8: Add Audit Trail for Permission Changes
-- ============================================================================

-- Create table for tracking permission assignment changes
CREATE TABLE IF NOT EXISTS permission_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL CHECK (action IN ('granted', 'revoked')),
    performed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    performed_at TIMESTAMP DEFAULT NOW(),
    reason TEXT
);

CREATE INDEX IF NOT EXISTS idx_permission_audit_log_tenant_role 
    ON permission_audit_log(tenant_id, role_id, performed_at DESC);

CREATE INDEX IF NOT EXISTS idx_permission_audit_log_permission 
    ON permission_audit_log(permission_id, performed_at DESC);

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================
