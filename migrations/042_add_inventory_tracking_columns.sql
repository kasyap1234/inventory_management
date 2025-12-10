-- Add reserved_quantity column to inventory table
-- This column was referenced in several places but not added to the schema

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'inventory' AND column_name = 'reserved_quantity'
    ) THEN
        ALTER TABLE inventory ADD COLUMN reserved_quantity INTEGER NOT NULL DEFAULT 0;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'inventory' AND column_name = 'minimum_level'
    ) THEN
        ALTER TABLE inventory ADD COLUMN minimum_level INTEGER DEFAULT 0;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'inventory' AND column_name = 'maximum_level'
    ) THEN
        ALTER TABLE inventory ADD COLUMN maximum_level INTEGER DEFAULT 0;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'inventory' AND column_name = 'reorder_point'
    ) THEN
        ALTER TABLE inventory ADD COLUMN reorder_point INTEGER DEFAULT 0;
    END IF;
END $$;
