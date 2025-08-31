-- Add category permissions to existing database
INSERT INTO permissions (name, description) VALUES
  ('categories:list', 'Can list categories'),
  ('categories:create', 'Can create new categories'),
  ('categories:read', 'Can read category information'),
  ('categories:update', 'Can update categories'),
  ('categories:delete', 'Can delete categories')
ON CONFLICT (name) DO NOTHING;

-- Get the user role ID for the default tenant
-- Assuming the user role exists, add category permissions to it
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id as role_id,
    p.id as permission_id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'user'
  AND r.tenant_id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid
  AND p.name IN ('categories:list', 'categories:read', 'categories:create')
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );