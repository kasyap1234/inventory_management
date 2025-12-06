-- Backfill critical permissions for tenant admins to unblock dashboard access
-- Ensures analytics and invoice permissions exist and are assigned to every admin role

-- 1) Ensure required permissions exist (dot notation)
INSERT INTO permissions (name, resource, action, description)
VALUES
    ('analytics.read', 'analytics', 'read', 'View analytics'),
    ('analytics.dashboard', 'analytics', 'dashboard', 'View dashboard analytics'),
    ('analytics.export', 'analytics', 'export', 'Export analytics data'),
    ('invoice.list', 'invoice', 'list', 'List all invoices'),
    ('invoice.create', 'invoice', 'create', 'Create invoices'),
    ('invoice.read', 'invoice', 'read', 'View invoice details'),
    ('invoice.update', 'invoice', 'update', 'Update invoice details'),
    ('invoice.delete', 'invoice', 'delete', 'Delete invoices'),
    ('invoice.bulk_create', 'invoice', 'bulk_create', 'Bulk create invoices'),
    ('invoice.update_status', 'invoice', 'update_status', 'Update invoice status'),
    ('invoice.generate_pdf', 'invoice', 'generate_pdf', 'Generate invoice PDF')
ON CONFLICT (name) DO NOTHING;

-- 2) Assign these permissions to every admin role across tenants
WITH target_permissions AS (
    SELECT id FROM permissions WHERE name IN (
        'analytics.read',
        'analytics.dashboard',
        'analytics.export',
        'invoice.list',
        'invoice.create',
        'invoice.read',
        'invoice.update',
        'invoice.delete',
        'invoice.bulk_create',
        'invoice.update_status',
        'invoice.generate_pdf'
    )
)
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, NOW()
FROM roles r
CROSS JOIN target_permissions p
WHERE r.name = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

