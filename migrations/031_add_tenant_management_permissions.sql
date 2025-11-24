-- Migration: Add tenant management permissions
-- These permissions control access to the multi-tenant system administration

-- 1. Add tenant management permissions
INSERT INTO permissions (name, resource, action, description) VALUES
('tenants:list', 'tenants', 'list', 'Can list all tenants in the system'),
('tenants:read', 'tenants', 'read', 'Can view tenant details'),
('tenants:create', 'tenants', 'create', 'Can create new tenants'),
('tenants:update', 'tenants', 'update', 'Can update tenant information'),
('tenants:delete', 'tenants', 'delete', 'Can delete tenants'),
('tenant:manage_settings', 'tenant', 'manage_settings', 'Can manage own tenant settings'),
('system:admin', 'system', 'admin', 'System administrator with full access')
ON CONFLICT (name) DO NOTHING;

-- 2. Assign system:admin permission to super_admin role
-- Note: This uses the default development tenant ID. In production, adjust accordingly.
DO $$
DECLARE
    super_admin_role_id UUID;
    system_admin_perm_id UUID;
BEGIN
    -- Get the super_admin role ID
    SELECT id INTO super_admin_role_id
    FROM roles
    WHERE name = 'super_admin'
    LIMIT 1;

    -- Get the system:admin permission ID
    SELECT id INTO system_admin_perm_id
    FROM permissions
    WHERE name = 'system:admin';

    -- Assign permission to role if both exist
    IF super_admin_role_id IS NOT NULL AND system_admin_perm_id IS NOT NULL THEN
        INSERT INTO role_permissions (role_id, permission_id)
        VALUES (super_admin_role_id, system_admin_perm_id)
        ON CONFLICT (role_id, permission_id) DO NOTHING;
    END IF;
END $$;

-- 3. Assign tenant:manage_settings to admin role
DO $$
DECLARE
    admin_role_id UUID;
    manage_settings_perm_id UUID;
BEGIN
    -- Get the admin role ID
    SELECT id INTO admin_role_id
    FROM roles
    WHERE name = 'admin'
    LIMIT 1;

    -- Get the tenant:manage_settings permission ID
    SELECT id INTO manage_settings_perm_id
    FROM permissions
    WHERE name = 'tenant:manage_settings';

    -- Assign permission to role if both exist
    IF admin_role_id IS NOT NULL AND manage_settings_perm_id IS NOT NULL THEN
        INSERT INTO role_permissions (role_id, permission_id)
        VALUES (admin_role_id, manage_settings_perm_id)
        ON CONFLICT (role_id, permission_id) DO NOTHING;
    END IF;
END $$;
