-- Fix RBAC permissions for product operations
-- This script assigns default product permissions to new users and ensures admin users have cross-tenant capabilities

-- Ensure basic product permissions exist
INSERT INTO permissions (name, description) VALUES
('read_products', 'Can read product information'),
('create_products', 'Can create new products'),
('update_products', 'Can update product information'),
('delete_products', 'Can delete products')
ON CONFLICT (name) DO NOTHING;

-- Add cross-tenant product permissions for admin users
INSERT INTO permissions (name, description) VALUES
('products:create_any_tenant', 'Can create products for any tenant'),
('products:read_any_tenant', 'Can read products for any tenant'),
('products:update_any_tenant', 'Can update products for any tenant'),
('products:delete_any_tenant', 'Can delete products for any tenant')
ON CONFLICT (name) DO NOTHING;

-- Assign product permissions to 'user' role for all tenants
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id as role_id, p.id as permission_id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'user'
  AND p.name IN ('read_products', 'create_products', 'update_products')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp
      WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- Assign cross-tenant product permissions to 'admin' role for all tenants
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id as role_id, p.id as permission_id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
  AND p.name LIKE 'products:%_any_tenant'
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp
      WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- Assign 'user' role to existing users who don't have any role assigned
-- Note: Works with or without tenant_id column (added in migration 020)
DO $$ 
DECLARE
    has_tenant_id boolean;
BEGIN
    -- Check if tenant_id column exists in user_roles
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_roles' AND column_name = 'tenant_id'
    ) INTO has_tenant_id;
    
    IF has_tenant_id THEN
        -- Insert with tenant_id if column exists
        INSERT INTO user_roles (user_id, role_id, tenant_id)
        SELECT u.id as user_id, r.id as role_id, u.tenant_id
        FROM users u
        JOIN roles r ON r.tenant_id = u.tenant_id AND r.name = 'user'
        WHERE NOT EXISTS (
            SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id AND ur.role_id = r.id
        );
    ELSE
        -- Insert without tenant_id if column doesn't exist yet
        INSERT INTO user_roles (user_id, role_id)
        SELECT u.id as user_id, r.id as role_id
        FROM users u
        JOIN roles r ON r.tenant_id = u.tenant_id AND r.name = 'user'
        WHERE NOT EXISTS (
            SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id AND ur.role_id = r.id
        );
    END IF;
END $$;