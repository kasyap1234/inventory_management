-- Insert sample categories for testing
-- First, get tenant ID (assuming there's at least one tenant)
DO $$
DECLARE
    tenant_id_var UUID;
BEGIN
    SELECT id INTO tenant_id_var FROM tenants LIMIT 1;
    IF tenant_id_var IS NULL THEN
        RAISE NOTICE 'No tenants found, cannot insert sample categories';
        RETURN;
    END IF;

    -- Insert root categories
    INSERT INTO categories (id, tenant_id, name, description, parent_id, level, path)
    VALUES
    (gen_random_uuid(), tenant_id_var, 'Fruits', 'Fresh fruits and produce', NULL, 0, 'Fruits'),
    (gen_random_uuid(), tenant_id_var, 'Vegetables', 'Fresh vegetables and greens', NULL, 0, 'Vegetables'),
    (gen_random_uuid(), tenant_id_var, 'Seeds', 'Crop seeds and planting materials', NULL, 0, 'Seeds'),
    (gen_random_uuid(), tenant_id_var, 'Fertilizers', 'Chemical and organic fertilizers', NULL, 0, 'Fertilizers'),
    (gen_random_uuid(), tenant_id_var, 'Equipment', 'Farming tools and machinery', NULL, 0, 'Equipment')
    ON CONFLICT (tenant_id, name) DO NOTHING;

    RAISE NOTICE 'Sample categories inserted for tenant %', tenant_id_var;
END $$;