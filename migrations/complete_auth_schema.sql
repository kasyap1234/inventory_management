-- Complete schema for authentication system with password hashing

-- Create tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(100) NOT NULL UNIQUE,
    license_number VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create users table with password hash
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255),  -- Added for password-based authentication
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

-- Create permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create user_roles table
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (user_id, role_id)
);

-- Create role_permissions table
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (role_id, permission_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_tenants_subdomain ON tenants (subdomain);
CREATE INDEX IF NOT EXISTS idx_users_tenant_email ON users (tenant_id, email);
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users (tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- Create default permissions
INSERT INTO permissions (name, description) VALUES
('read_users', 'Can read user information'),
('create_users', 'Can create new users'),
('update_users', 'Can update user information'),
('delete_users', 'Can delete users'),
('read_products', 'Can read product information'),
('create_products', 'Can create new products'),
('update_products', 'Can update product information'),
('delete_products', 'Can delete products'),
('manage_products', 'Can perform bulk product operations'),
('tenants:list', 'Can list tenants'),
('tenants:create', 'Can create new tenants'),
('tenants:read', 'Can read tenant information'),
('tenants:update', 'Can update tenant information'),
('tenants:delete', 'Can delete tenants'),
('users:list', 'Can list users'),
('users:read', 'Can read user information'),
('users:create', 'Can create new users'),
('users:update', 'Can update user information'),
('users:delete', 'Can delete users')
ON CONFLICT (name) DO NOTHING;

-- Create a default tenant
INSERT INTO tenants (id, name, subdomain, license_number, status)
VALUES (
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid,
    'Agromart Development',
    'agromart-dev',
    'DEV001',
    'active'
) ON CONFLICT DO NOTHING;

-- Create default roles
INSERT INTO roles (id, tenant_id, name, description) VALUES
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid, 'admin', 'Administrator role'),
('cccccccc-cccc-cccc-cccc-cccccccccccc'::uuid, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid, 'user', 'Regular user role')
ON CONFLICT DO NOTHING;

-- Link admin role to all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid as role_id,
    p.id as permission_id
FROM permissions p
WHERE NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid
    AND rp.permission_id = p.id
);

-- Link user role to basic read permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    'cccccccc-cccc-cccc-cccc-cccccccccccc'::uuid as role_id,
    p.id as permission_id
FROM permissions p
WHERE p.name LIKE 'read_%'
AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = 'cccccccc-cccc-cccc-cccc-cccccccccccc'::uuid
    AND rp.permission_id = p.id
);