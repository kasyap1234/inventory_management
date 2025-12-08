-- Ensure admin roles always have full access
-- 1) Create a wildcard permission (*) if it doesn't exist
INSERT INTO permissions (name, resource, action, description)
VALUES ('*', 'system', 'all', 'Wildcard permission granting full access')
ON CONFLICT (name) DO NOTHING;

-- 2) Assign the wildcard permission to all admin and super_admin roles
WITH wildcard AS (
    SELECT id FROM permissions WHERE name = '*'
),
target_roles AS (
    SELECT id
    FROM roles
    WHERE name IN ('admin', 'super_admin')
      AND is_active = true
)
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT tr.id, w.id, NOW()
FROM target_roles tr
CROSS JOIN wildcard w
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 3) Backfill any missing concrete permissions to admin and super_admin roles
WITH target_roles AS (
    SELECT id
    FROM roles
    WHERE name IN ('admin', 'super_admin')
      AND is_active = true
),
all_perms AS (
    SELECT id FROM permissions
)
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT tr.id, p.id, NOW()
FROM target_roles tr
CROSS JOIN all_perms p
ON CONFLICT (role_id, permission_id) DO NOTHING;

