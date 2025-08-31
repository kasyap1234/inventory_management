-- Fix user tenant_id constraints to prevent NULL and NIL UUID

ALTER TABLE users ALTER COLUMN tenant_id SET NOT NULL;

-- Add constraint to prevent NIL UUID (zero UUID)

ALTER TABLE users ADD CONSTRAINT chk_tenant_id_not_nil CHECK (tenant_id != '00000000-0000-0000-0000-000000000000'::uuid);

-- Add comment for clarity

COMMENT ON COLUMN users.tenant_id IS 'Tenant ID - must be non-null and not nil UUID';