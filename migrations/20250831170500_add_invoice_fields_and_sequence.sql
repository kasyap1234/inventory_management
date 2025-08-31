-- Add invoice numbering sequence and new columns to invoices table
-- Migration: 20250831170500_add_invoice_fields_and_sequence.sql

-- Create sequence for invoice numbering
CREATE SEQUENCE IF NOT EXISTS invoice_number_seq START WITH 1;

-- Add new columns to invoices table
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS invoice_number VARCHAR(50) NOT NULL DEFAULT '';
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS paid_date TIMESTAMPTZ NULL;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS due_date DATE NOT NULL DEFAULT CURRENT_DATE;

-- Add index for invoice_number for faster lookups
CREATE INDEX IF NOT EXISTS idx_invoices_invoice_number ON invoices(invoice_number);
CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(due_date);
CREATE INDEX IF NOT EXISTS idx_invoices_status_due_date ON invoices(status, due_date);

-- Update existing invoices to have invoice numbers (if any exist)
-- Use tenant_id for tenant-specific numbering
UPDATE invoices SET
    invoice_number = CONCAT('INV-', tenant_id::text, '-', LPAD(EXTRACT(YEAR FROM issued_date)::text, 4, '0'), '-', LPAD(EXTRACT(MONTH FROM issued_date)::text, 2, '0'), '-', LPAD(id::text, 8, '0')),
    due_date = (issued_date + INTERVAL '30 days')::date
WHERE invoice_number = '';

-- Add constraint to ensure invoice status is one of the valid values including 'cancelled'
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoices_status_check;
ALTER TABLE invoices ADD CONSTRAINT invoices_status_check
    CHECK (status IN ('unpaid', 'paid', 'overdue', 'cancelled'));

-- Create invoice_sequences table for per-tenant invoice number tracking
CREATE TABLE IF NOT EXISTS invoice_sequences (
    tenant_id UUID NOT NULL,
    year_month VARCHAR(7) NOT NULL, -- Format: YYYY-MM
    last_number INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, year_month)
);

-- Add index for faster sequence lookups
CREATE INDEX IF NOT EXISTS idx_invoice_sequences_tenant_year_month ON invoice_sequences(tenant_id, year_month);