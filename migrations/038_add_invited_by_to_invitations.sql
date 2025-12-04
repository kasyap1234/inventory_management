-- Add missing invited_by column to invitations table
ALTER TABLE invitations ADD COLUMN IF NOT EXISTS invited_by UUID REFERENCES users(id) ON DELETE SET NULL;
