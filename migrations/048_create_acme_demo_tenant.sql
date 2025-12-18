-- Create Acme demo tenant and user
-- This migration creates the Acme tenant so that 041_seed_acme_sample_data.sql can populate demo data

DO $$
DECLARE
    v_tenant_id   CONSTANT uuid := 'ff5c2ea7-4d2d-4928-a34f-faac6be42c71';
    v_user_id     CONSTANT uuid := '812ae9d8-6a2d-4382-a0ca-bacf044613d6';
    v_admin_role  CONSTANT uuid := 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb';
    -- Password: AcmeDemo123!
    -- Generated with: SELECT crypt('AcmeDemo123!', gen_salt('bf', 10));
    v_password_hash TEXT;
BEGIN
    -- Generate password hash for 'AcmeDemo123!'
    v_password_hash := crypt('AcmeDemo123!', gen_salt('bf', 10));

    -- Create Acme tenant if not exists
    INSERT INTO tenants (id, name, subdomain, status, created_at, updated_at)
    VALUES (v_tenant_id, 'Acme Corporation', 'acme', 'active', NOW(), NOW())
    ON CONFLICT (id) DO NOTHING;

    -- Create Acme admin user if not exists
    INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, status, created_at, updated_at)
    VALUES (v_user_id, v_tenant_id, 'acme@gmail.com', v_password_hash, 'Acme', 'Admin', 'active', NOW(), NOW())
    ON CONFLICT (id) DO UPDATE SET
        password_hash = EXCLUDED.password_hash,
        updated_at = NOW();

    -- Assign admin role to Acme user
    INSERT INTO user_roles (id, user_id, role_id, tenant_id, created_at)
    VALUES (gen_random_uuid(), v_user_id, v_admin_role, v_tenant_id, NOW())
    ON CONFLICT DO NOTHING;

    RAISE NOTICE 'Acme tenant and user created. Login: acme@gmail.com / AcmeDemo123!';
END $$;
