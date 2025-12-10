-- Migration: Add RBAC Role Management and Missing Permissions
-- This migration adds permissions for role management, billing, and other missing modules
-- to enable fine-grained access control

-- Role Management permissions
INSERT INTO permissions (name, resource, action, description) VALUES
-- Role CRUD operations
('role.list', 'role', 'list', 'List all roles'),
('role.create', 'role', 'create', 'Create new roles'),
('role.read', 'role', 'read', 'View role details'),
('role.update', 'role', 'update', 'Update role information'),
('role.delete', 'role', 'delete', 'Delete roles'),

-- Role permission management
('role.manage_permissions', 'role', 'manage_permissions', 'Assign or remove permissions from roles'),
('role.assign_users', 'role', 'assign_users', 'Assign users to roles'),
('role.remove_users', 'role', 'remove_users', 'Remove users from roles'),

-- Billing and Subscription permissions
('billing.read', 'billing', 'read', 'View billing information'),
('billing.update', 'billing', 'update', 'Update billing settings'),
('billing.manage_payment_methods', 'billing', 'manage_payment_methods', 'Manage payment methods'),
('subscription.read', 'subscription', 'read', 'View subscription details'),
('subscription.update', 'subscription', 'update', 'Update subscription plan'),
('subscription.cancel', 'subscription', 'cancel', 'Cancel subscription'),

-- Purchasing/Order receiving permissions (more granular)
('order.receive', 'order', 'receive', 'Receive purchase orders'),
('order.approve', 'order', 'approve', 'Approve orders'),
('order.process', 'order', 'process', 'Process orders'),
('order.ship', 'order', 'ship', 'Ship orders'),
('order.deliver', 'order', 'deliver', 'Mark orders as delivered'),
('order.cancel', 'order', 'cancel', 'Cancel orders'),

-- Reports and Analytics (more granular)
('reports.read', 'reports', 'read', 'View reports'),
('reports.export', 'reports', 'export', 'Export reports'),
('reports.create', 'reports', 'create', 'Create custom reports'),
('analytics.dashboard', 'analytics', 'dashboard', 'View dashboard analytics'),
('analytics.export', 'analytics', 'export', 'Export analytics data'),

-- Warehouse management (already exists but ensuring completeness)
('warehouse.list', 'warehouse', 'list', 'List warehouses'),
('warehouse.create', 'warehouse', 'create', 'Create warehouses'),
('warehouse.read', 'warehouse', 'read', 'View warehouse details'),
('warehouse.update', 'warehouse', 'update', 'Update warehouses'),
('warehouse.delete', 'warehouse', 'delete', 'Delete warehouses'),

-- Category management (already exists but ensuring completeness)
('category.list', 'category', 'list', 'List categories'),
('category.create', 'category', 'create', 'Create categories'),
('category.read', 'category', 'read', 'View category details'),
('category.update', 'category', 'update', 'Update categories'),
('category.delete', 'category', 'delete', 'Delete categories'),

-- Invoice management (more granular)
('invoice.update_status', 'invoice', 'update_status', 'Update invoice status'),
('invoice.generate_pdf', 'invoice', 'generate_pdf', 'Generate invoice PDF'),
('invoice.bulk_create', 'invoice', 'bulk_create', 'Bulk create invoices')

ON CONFLICT (name) DO UPDATE SET
    resource = EXCLUDED.resource,
    action = EXCLUDED.action,
    description = EXCLUDED.description,
    updated_at = NOW();

-- Add comment for documentation
COMMENT ON TABLE permissions IS 'System permissions - includes role management, billing, and fine-grained module permissions as of migration 046';
