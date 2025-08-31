-- Migration to create complete business tables schema
-- This replaces and completes the missing tables that are expected by the application

-- Supporting Tables for Business Logic

-- Warehouses
CREATE TABLE IF NOT EXISTS warehouses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  address TEXT,
  capacity INTEGER CHECK (capacity > 0),
  license_number VARCHAR(100),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  UNIQUE (tenant_id, name)
);

-- Suppliers
CREATE TABLE IF NOT EXISTS suppliers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  contact_email VARCHAR(255),
  contact_phone VARCHAR(20),
  address TEXT,
  license_number VARCHAR(100),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Distributors
CREATE TABLE IF NOT EXISTS distributors (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  contact_email VARCHAR(255),
  contact_phone VARCHAR(20),
  address TEXT,
  license_number VARCHAR(100),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Inventory
CREATE TABLE IF NOT EXISTS inventory (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
  last_updated TIMESTAMP DEFAULT NOW(),
  UNIQUE (tenant_id, warehouse_id, product_id)
);

-- Orders
CREATE TABLE IF NOT EXISTS orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  order_type VARCHAR(10) NOT NULL CHECK (order_type IN ('purchase', 'sales')),
  supplier_id UUID REFERENCES suppliers(id),
  distributor_id UUID REFERENCES distributors(id),
  product_id UUID NOT NULL REFERENCES products(id),
  warehouse_id UUID NOT NULL REFERENCES warehouses(id),
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  unit_price DECIMAL(10,2) NOT NULL,
  status VARCHAR(50) DEFAULT 'pending' NOT NULL CHECK (status IN ('pending', 'approved', 'received', 'shipped', 'delivered', 'cancelled')),
  order_date DATE DEFAULT NOW(),
  expected_delivery DATE,
  notes TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  CHECK (order_type = 'purchase' AND supplier_id IS NOT NULL AND distributor_id IS NULL OR
    order_type = 'sales' AND distributor_id IS NOT NULL AND supplier_id IS NULL)
);

-- Invoices
CREATE TABLE IF NOT EXISTS invoices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  gstin VARCHAR(15),
  hsn_sac VARCHAR(6),
  taxable_amount DECIMAL(10,2),
  gst_rate DECIMAL(5,2),
  cgst DECIMAL(5,2),
  sgst DECIMAL(5,2),
  igst DECIMAL(5,2),
  total_amount DECIMAL(10,2) NOT NULL,
  status VARCHAR(50) DEFAULT 'unpaid' NOT NULL CHECK (status IN ('unpaid', 'paid', 'overdue', 'cancelled')),
  issued_date DATE DEFAULT NOW(),
  paid_date DATE,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Product Images
CREATE TABLE IF NOT EXISTS product_images (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  image_url VARCHAR(500) NOT NULL,
  alt_text VARCHAR(100),
  created_at TIMESTAMP DEFAULT NOW()
);

-- Subscriptions
CREATE TABLE IF NOT EXISTS subscriptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  razorpay_subscription_id VARCHAR(255) UNIQUE,
  plan_name VARCHAR(100) NOT NULL,
  amount DECIMAL(10,2) NOT NULL,
  currency VARCHAR(3) DEFAULT 'INR',
  billing_frequency VARCHAR(20) DEFAULT 'monthly' CHECK (billing_frequency IN ('monthly', 'quarterly', 'annually')),
  status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'cancelled', 'suspended')),
  current_period_start TIMESTAMP,
  current_period_end TIMESTAMP,
  start_date DATE NOT NULL,
  end_date DATE,
  razorpay_started_at TIMESTAMP,
  razorpay_ends_at TIMESTAMP,
  cancelled_at TIMESTAMP,
  notes TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id),
  title VARCHAR(255) NOT NULL,
  message TEXT NOT NULL,
  notification_type VARCHAR(50) DEFAULT 'info' CHECK (notification_type IN ('info', 'warning', 'error', 'success')),
  priority VARCHAR(20) DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'critical')),
  status VARCHAR(20) DEFAULT 'unread' CHECK (status IN ('unread', 'read', 'archived')),
  read_at TIMESTAMP,
  expires_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Audit Logs
CREATE TABLE IF NOT EXISTS audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  table_name VARCHAR(50) NOT NULL,
  record_id VARCHAR(100) NOT NULL,
  action VARCHAR(20) NOT NULL,
  new_values JSONB,
  old_values JSONB,
  changed_by UUID REFERENCES users(id),
  deleted BOOLEAN DEFAULT FALSE,
  deleted_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW()
) PARTITION BY HASH (tenant_id);

-- Example partition (create as needed for scaling)
CREATE TABLE IF NOT EXISTS audit_logs_p0 PARTITION OF audit_logs DEFAULT;

-- Tokens (for session management)
CREATE TABLE IF NOT EXISTS tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id),
  token VARCHAR(500) NOT NULL UNIQUE,
  token_type VARCHAR(50) DEFAULT 'session' CHECK (token_type IN ('session', 'reset', 'verification')),
  expires_at TIMESTAMP NOT NULL,
  used BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_warehouses_tenant_name ON warehouses (tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_suppliers_tenant_name ON suppliers (tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_distributors_tenant_name ON distributors (tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_warehouse_product ON inventory (tenant_id, warehouse_id, product_id);
CREATE INDEX IF NOT EXISTS idx_inventory_tenant_last_updated ON inventory (tenant_id, last_updated);
CREATE INDEX IF NOT EXISTS idx_orders_tenant_status ON orders (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_orders_tenant_date ON orders (tenant_id, order_date);
CREATE INDEX IF NOT EXISTS idx_orders_tenant_type ON orders (tenant_id, order_type);
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_status ON invoices (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_product_images_tenant_product ON product_images (tenant_id, product_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant_status ON subscriptions (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant_end_date ON subscriptions (tenant_id, end_date);
CREATE INDEX IF NOT EXISTS idx_notifications_tenant_user_status ON notifications (tenant_id, user_id, status);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_created ON audit_logs (tenant_id, created_at);
CREATE INDEX IF NOT EXISTS idx_tokens_tenant_expires ON tokens (tenant_id, expires_at);
CREATE INDEX IF NOT EXISTS idx_tokens_user_token ON tokens (user_id, token);

-- Add sample data for business operations
-- (Can be removed or commented out for production)
INSERT INTO warehouses (tenant_id, name, address, capacity, license_number) VALUES
((SELECT id FROM tenants LIMIT 1), 'Main Warehouse', '123 Industrial Area, Mumbai', 10000, 'WH001');

INSERT INTO suppliers (tenant_id, name, contact_email, contact_phone, address, license_number) VALUES
((SELECT id FROM tenants LIMIT 1), 'Agri Supplies Inc', 'contact@agrisupplies.com', '+91-9876543210', '456 Supplier Lane, Delhi', 'SUP001');

INSERT INTO distributors (tenant_id, name, contact_email, contact_phone, address, license_number) VALUES
((SELECT id FROM tenants LIMIT 1), 'Farm Distribution Co', 'sales@farmdistro.com', '+91-9876543211', '789 Distributor Road, Kolkata', 'DIST001');

INSERT INTO inventory (tenant_id, warehouse_id, product_id, quantity, last_updated) VALUES
((SELECT id FROM tenants LIMIT 1), (SELECT id FROM warehouses LIMIT 1), (SELECT id FROM products LIMIT 1), 500, NOW());