-- Add contact information fields to tenants table
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS contact_email VARCHAR(255);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS contact_phone VARCHAR(50);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS support_email VARCHAR(255);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS support_phone VARCHAR(50);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS address TEXT;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS city VARCHAR(100);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS state VARCHAR(100);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS country VARCHAR(100);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS postal_code VARCHAR(20);

-- Add indexes for commonly queried fields
CREATE INDEX IF NOT EXISTS idx_tenants_contact_email ON tenants (contact_email);
CREATE INDEX IF NOT EXISTS idx_tenants_support_email ON tenants (support_email);

-- Add comment to describe the purpose
COMMENT ON COLUMN tenants.contact_email IS 'Primary contact email for the tenant organization';
COMMENT ON COLUMN tenants.contact_phone IS 'Primary contact phone for the tenant organization';
COMMENT ON COLUMN tenants.support_email IS 'Support email displayed on invoices and customer communications';
COMMENT ON COLUMN tenants.support_phone IS 'Support phone displayed on invoices and customer communications';
