-- Add warehouse permissions to existing database
INSERT INTO permissions (name, description) VALUES
  ('warehouses:list', 'Can list warehouses'),
  ('warehouses:create', 'Can create new warehouses'),
  ('warehouses:read', 'Can read warehouse information'),
  ('warehouses:update', 'Can update warehouses'),
  ('warehouses:delete', 'Can delete warehouses')
ON CONFLICT (name) DO NOTHING;

-- Add warehouse permissions to all 'user' roles across all tenants
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id as role_id,
    p.id as permission_id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'user'
  AND p.name IN ('warehouses:list', 'warehouses:read', 'warehouses:create')
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );