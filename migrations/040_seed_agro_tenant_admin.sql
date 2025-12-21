-- Migration: Seed Acme Tenant Admin for Demo
-- This script ensures the Acme tenant and its admin exist before rich data is seeded.
-- Idempotent and safe for cross-environment setup.

DO $$
DECLARE
    v_tenant_id UUID := 'ff5c2ea7-4d2d-4928-a34f-faac6be42c71';
    v_user_id   UUID := '812ae9d8-6a2d-4382-a0ca-bacf044613d6';
    v_role_id   UUID := '912ae9d8-6a2d-4382-a0ca-bacf044613e7';
    v_email     VARCHAR := 'acme@gmail.com';
    v_pass_hash VARCHAR := '$2a$10$5h1jeuCqdExRm5zlffUwBejcVQsjskCkFWZcx3tBHIlOw545OpG1a'; -- Acme@12345!
BEGIN
    -- 1. Insert Tenant
    INSERT INTO tenants (id, name, subdomain, license_number, status)
    VALUES (v_tenant_id, 'Acme Corp', 'acme', 'ACME-DEMO-001', 'active')
    ON CONFLICT (id) DO NOTHING;
    
    -- Ensure subdomain matches if record existed with different ID
    UPDATE tenants SET subdomain = 'acme' WHERE id = v_tenant_id;

    -- 2. Ensure tenant_admin role exists for this tenant
    INSERT INTO roles (id, tenant_id, name, description)
    VALUES (v_role_id, v_tenant_id, 'tenant_admin', 'Full Administrator for Acme Corp')
    ON CONFLICT (tenant_id, name) DO UPDATE SET description = EXCLUDED.description
    RETURNING id INTO v_role_id;

    -- 3. Link ALL current permissions to this tenant_admin role
    INSERT INTO role_permissions (role_id, permission_id)
    SELECT v_role_id, id FROM permissions
    ON CONFLICT (role_id, permission_id) DO NOTHING;

    -- 4. Create User
    INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, status)
    VALUES (v_user_id, v_tenant_id, v_email, v_pass_hash, 'Acme', 'Admin', 'active')
    ON CONFLICT (tenant_id, email) DO UPDATE SET 
        password_hash = EXCLUDED.password_hash,
        status = EXCLUDED.status;

    -- 5. Assign role to user
    INSERT INTO user_roles (user_id, tenant_id, role_id)
    VALUES (v_user_id, v_tenant_id, v_role_id)
    ON CONFLICT (user_id, tenant_id, role_id) DO NOTHING;

    RAISE NOTICE 'Seed Admin for % created/verified.', v_email;
END $$;
