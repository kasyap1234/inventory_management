-- Add all business permissions to existing database
INSERT INTO permissions (name, description) VALUES
  -- Suppliers permissions
  ('suppliers:list', 'Can list suppliers'),
  ('suppliers:create', 'Can create new suppliers'),
  ('suppliers:read', 'Can read supplier information'),
  ('suppliers:update', 'Can update suppliers'),
  ('suppliers:delete', 'Can delete suppliers'),

  -- Inventory permissions
  ('inventories:list', 'Can list inventory'),
  ('inventories:create', 'Can create inventory entries'),
  ('inventories:read', 'Can read inventory information'),
  ('inventories:update', 'Can update inventory'),
  ('inventories:delete', 'Can delete inventory'),

  -- Orders permissions
  ('orders:list', 'Can list orders'),
  ('orders:create', 'Can create new orders'),
  ('orders:read', 'Can read order information'),
  ('orders:update', 'Can update orders'),
  ('orders:delete', 'Can delete orders'),

  -- Invoices permissions
  ('invoices:list', 'Can list invoices'),
  ('invoices:create', 'Can create new invoices'),
  ('invoices:read', 'Can read invoice information'),
  ('invoices:update', 'Can update invoices'),
  ('invoices:delete', 'Can delete invoices')
ON CONFLICT (name) DO NOTHING;

-- Get the user role ID for the default tenant
-- Add all business permissions to the user role
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id as role_id,
    p.id as permission_id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'user'
  AND r.tenant_id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid
  AND p.name LIKE '%:%'
  AND (p.name LIKE 'warehouses:%' OR p.name LIKE 'distributors:%' OR
       p.name LIKE 'suppliers:%' OR p.name LIKE 'inventories:%' OR
       p.name LIKE 'orders:%' OR p.name LIKE 'invoices:%')
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );