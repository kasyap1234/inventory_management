-- Add distributors permissions to existing database
INSERT INTO permissions (name, description) VALUES
  ('distributors:list', 'Can list distributors'),
  ('distributors:create', 'Can create new distributors'),
  ('distributors:read', 'Can read distributor information'),
  ('distributors:update', 'Can update distributors'),
  ('distributors:delete', 'Can delete distributors')
ON CONFLICT (name) DO NOTHING;

-- Add distributor permissions to all 'user' roles across all tenants
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id as role_id,
    p.id as permission_id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'user'
  AND p.name IN ('distributors:list', 'distributors:read', 'distributors:create')
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );