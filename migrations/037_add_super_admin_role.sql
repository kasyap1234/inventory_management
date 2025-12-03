-- Add Super Admin and Tenant Admin roles

-- 1. Create a "System" tenant if it doesn't exist (for Super Admins)
-- We'll use a fixed UUID or look it up. For now, let's just ensure roles exist.
-- Ideally, Super Admin is a global concept, but our system is multi-tenant.
-- We can treat the "System" tenant as a special tenant.

-- 2. Insert 'super_admin' and 'tenant_admin' roles into the roles table.
-- Since roles are tenant-scoped, we might need to add these roles when a tenant is created.
-- However, 'super_admin' might need to be in a specific "System" tenant.
-- For 'tenant_admin', it's a template role that should be available/created for every tenant.

-- Let's define the permissions for these roles first.

-- Insert new permissions for tenant management
INSERT INTO permissions (name, resource, action, description) VALUES
('tenant.invite', 'tenant', 'invite', 'Invite new tenants (Super Admin only)'),
('tenant.manage_users', 'tenant', 'manage_users', 'Manage users within the tenant')
ON CONFLICT (name) DO NOTHING;

-- We don't insert roles here because roles are tenant-specific.
-- Instead, we will rely on the application logic (auth_service.go) to create these roles
-- when a new tenant is created (for tenant_admin) or when the system initializes (for super_admin).

-- However, we can create a "template" or "definition" if we had a global roles table.
-- Since we don't, this migration will primarily ensure the permissions exist.
