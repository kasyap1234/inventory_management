-- Migration to add category hierarchy support
-- Add parent_id, level, and path columns to categories table

ALTER TABLE categories ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES categories(id);
ALTER TABLE categories ADD COLUMN IF NOT EXISTS level INTEGER DEFAULT 0;
ALTER TABLE categories ADD COLUMN IF NOT EXISTS path TEXT;

-- Update path for existing categories to their name (or ID if name is null)
UPDATE categories SET path = COALESCE(name, id::text), level = 0 WHERE path IS NULL;

-- Create partial index for performance
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories (tenant_id, parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_categories_level ON categories (tenant_id, level);
CREATE INDEX IF NOT EXISTS idx_categories_path ON categories (tenant_id, path);