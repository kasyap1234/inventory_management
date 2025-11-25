-- Standardize Permission Naming Convention
-- This migration adds standardized permissions using dot notation (resource.action)
-- to replace the inconsistent mix of colon and dot notation currently in use.
-- Old permissions are kept for backward compatibility during the transition period.

-- Insert standardized User Management permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('user.list', 'user', 'list', 'List all users'),
('user.create', 'user', 'create', 'Create new users'),
('user.read', 'user', 'read', 'View user information'),
('user.update', 'user', 'update', 'Update user information'),
('user.delete', 'user', 'delete', 'Delete users'),
('user.approve', 'user', 'approve', 'Approve pending users'),
('user.create_any_tenant', 'user', 'create_any_tenant', 'Create users in any tenant')
ON CONFLICT (name) DO NOTHING;

-- Insert standardized Product Management permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('product.list', 'product', 'list', 'List all products'),
('product.create', 'product', 'create', 'Create new products'),
('product.read', 'product', 'read', 'View product information'),
('product.update', 'product', 'update', 'Update product information'),
('product.delete', 'product', 'delete', 'Delete products')
ON CONFLICT (name) DO NOTHING;

-- Insert standardized Order Management permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('order.list', 'order', 'list', 'List all orders'),
('order.create', 'order', 'create', 'Create new orders'),
('order.read', 'order', 'read', 'View order information'),
('order.update', 'order', 'update', 'Update order information'),
('order.delete', 'order', 'delete', 'Delete orders')
ON CONFLICT (name) DO NOTHING;

-- Insert standardized Invoice Management permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('invoice.list', 'invoice', 'list', 'List all invoices'),
('invoice.create', 'invoice', 'create', 'Create new invoices'),
('invoice.read', 'invoice', 'read', 'View invoice information'),
('invoice.update', 'invoice', 'update', 'Update invoice information'),
('invoice.delete', 'invoice', 'delete', 'Delete invoices')
ON CONFLICT (name) DO NOTHING;

-- Insert standardized Inventory Management permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('inventory.list', 'inventory', 'list', 'List all inventory items'),
('inventory.create', 'inventory', 'create', 'Create new inventory items'),
('inventory.read', 'inventory', 'read', 'View inventory information'),
('inventory.update', 'inventory', 'update', 'Update inventory information'),
('inventory.delete', 'inventory', 'delete', 'Delete inventory items')
ON CONFLICT (name) DO NOTHING;

-- Insert standardized Supplier Management permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('supplier.list', 'supplier', 'list', 'List all suppliers'),
('supplier.create', 'supplier', 'create', 'Create new suppliers'),
('supplier.read', 'supplier', 'read', 'View supplier information'),
('supplier.update', 'supplier', 'update', 'Update supplier information'),
('supplier.delete', 'supplier', 'delete', 'Delete suppliers')
ON CONFLICT (name) DO NOTHING;

-- Insert standardized Tenant Management permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('tenant.list', 'tenant', 'list', 'List all tenants'),
('tenant.create', 'tenant', 'create', 'Create new tenants'),
('tenant.read', 'tenant', 'read', 'View tenant information'),
('tenant.update', 'tenant', 'update', 'Update tenant information'),
('tenant.delete', 'tenant', 'delete', 'Delete tenants')
ON CONFLICT (name) DO NOTHING;

-- Insert standardized Analytics permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('analytics.read', 'analytics', 'read', 'View analytics and reports')
ON CONFLICT (name) DO NOTHING;

-- Insert standardized Audit permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('audit.read', 'audit', 'read', 'View audit logs')
ON CONFLICT (name) DO NOTHING;

-- Insert standardized Batch Management permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('batch.list', 'batch', 'list', 'List all batches'),
('batch.create', 'batch', 'create', 'Create new batches'),
('batch.read', 'batch', 'read', 'View batch information'),
('batch.update', 'batch', 'update', 'Update batch information'),
('batch.delete', 'batch', 'delete', 'Delete batches')
ON CONFLICT (name) DO NOTHING;

-- Copy role_permission assignments from old colon notation to new dot notation
-- This ensures all existing roles retain their permissions with the new naming

-- User permissions mapping
INSERT INTO role_permissions (role_id, permission_id, conditions, granted_at)
SELECT DISTINCT rp.role_id, pnew.id, rp.conditions, NOW()
FROM role_permissions rp
JOIN permissions pold ON rp.permission_id = pold.id
JOIN permissions pnew ON pnew.name = 
    CASE pold.name
        WHEN 'users:list' THEN 'user.list'
        WHEN 'users:create' THEN 'user.create'
        WHEN 'users:read' THEN 'user.read'
        WHEN 'users:update' THEN 'user.update'
        WHEN 'users:delete' THEN 'user.delete'
        WHEN 'users:approve' THEN 'user.approve'
        WHEN 'users:create_any_tenant' THEN 'user.create_any_tenant'
        ELSE NULL
    END
WHERE pold.name IN ('users:list', 'users:create', 'users:read', 'users:update', 'users:delete', 'users:approve', 'users:create_any_tenant')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Inventory permissions mapping
INSERT INTO role_permissions (role_id, permission_id, conditions, granted_at)
SELECT DISTINCT rp.role_id, pnew.id, rp.conditions, NOW()
FROM role_permissions rp
JOIN permissions pold ON rp.permission_id = pold.id
JOIN permissions pnew ON pnew.name = 
    CASE pold.name
        WHEN 'inventories:list' THEN 'inventory.list'
        WHEN 'inventories:create' THEN 'inventory.create'
        WHEN 'inventories:read' THEN 'inventory.read'
        WHEN 'inventories:update' THEN 'inventory.update'
        WHEN 'inventories:delete' THEN 'inventory.delete'
        ELSE NULL
    END
WHERE pold.name IN ('inventories:list', 'inventories:create', 'inventories:read', 'inventories:update', 'inventories:delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Supplier permissions mapping
INSERT INTO role_permissions (role_id, permission_id, conditions, granted_at)
SELECT DISTINCT rp.role_id, pnew.id, rp.conditions, NOW()
FROM role_permissions rp
JOIN permissions pold ON rp.permission_id = pold.id
JOIN permissions pnew ON pnew.name = 
    CASE pold.name
        WHEN 'suppliers:list' THEN 'supplier.list'
        WHEN 'suppliers:create' THEN 'supplier.create'
        WHEN 'suppliers:read' THEN 'supplier.read'
        WHEN 'suppliers:update' THEN 'supplier.update'
        WHEN 'suppliers:delete' THEN 'supplier.delete'
        ELSE NULL
    END
WHERE pold.name IN ('suppliers:list', 'suppliers:create', 'suppliers:read', 'suppliers:update', 'suppliers:delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Distributor permissions mapping
INSERT INTO role_permissions (role_id, permission_id, conditions, granted_at)
SELECT DISTINCT rp.role_id, pnew.id, rp.conditions, NOW()
FROM role_permissions rp
JOIN permissions pold ON rp.permission_id = pold.id
JOIN permissions pnew ON pnew.name = 
    CASE pold.name
        WHEN 'distributors:list' THEN 'distributor.list'
        WHEN 'distributors:create' THEN 'distributor.create'
        WHEN 'distributors:read' THEN 'distributor.read'
        WHEN 'distributors:update' THEN 'distributor.update'
        WHEN 'distributors:delete' THEN 'distributor.delete'
        ELSE NULL
    END
WHERE pold.name IN ('distributors:list', 'distributors:create', 'distributors:read', 'distributors:update', 'distributors:delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Tenant permissions mapping
INSERT INTO role_permissions (role_id, permission_id, conditions, granted_at)
SELECT DISTINCT rp.role_id, pnew.id, rp.conditions, NOW()
FROM role_permissions rp
JOIN permissions pold ON rp.permission_id = pold.id
JOIN permissions pnew ON pnew.name = 
    CASE pold.name
        WHEN 'tenants:list' THEN 'tenant.list'
        WHEN 'tenants:read' THEN 'tenant.read'
        WHEN 'tenants:update' THEN 'tenant.update'
        WHEN 'tenants:delete' THEN 'tenant.delete'
        ELSE NULL
    END
WHERE pold.name IN ('tenants:list', 'tenants:read', 'tenants:update', 'tenants:delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Analytics permissions mapping
INSERT INTO role_permissions (role_id, permission_id, conditions, granted_at)
SELECT DISTINCT rp.role_id, pnew.id, rp.conditions, NOW()
FROM role_permissions rp
JOIN permissions pold ON rp.permission_id = pold.id
JOIN permissions pnew ON pnew.name = 'analytics.read'
WHERE pold.name = 'analytics:read'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Audit permissions mapping
INSERT INTO role_permissions (role_id, permission_id, conditions, granted_at)
SELECT DISTINCT rp.role_id, pnew.id, rp.conditions, NOW()
FROM role_permissions rp
JOIN permissions pold ON rp.permission_id = pold.id
JOIN permissions pnew ON pnew.name = 'audit.read'
WHERE pold.name = 'audit:read'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Add comment for documentation
COMMENT ON TABLE permissions IS 'System permissions - standardized to use dot notation (resource.action) as of migration 032';
