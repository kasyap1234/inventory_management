-- Enhanced Product Images Table
-- This migration creates a new table for enhanced product image management with multiple sizes

-- Create enhanced product images table
CREATE TABLE IF NOT EXISTS enhanced_product_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    original_url VARCHAR(1000) NOT NULL,
    large_url VARCHAR(1000),
    medium_url VARCHAR(1000),
    small_url VARCHAR(1000),
    thumbnail_url VARCHAR(1000),
    alt_text VARCHAR(255),
    file_size BIGINT NOT NULL CHECK (file_size > 0),
    mime_type VARCHAR(100) NOT NULL,
    width INTEGER NOT NULL CHECK (width > 0),
    height INTEGER NOT NULL CHECK (height > 0),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Add constraints (safely checking if they already exist)
DO $$ BEGIN
    -- Add mime_type constraint if it doesn't exist
    BEGIN
        ALTER TABLE enhanced_product_images 
        ADD CONSTRAINT chk_enhanced_product_images_mime_type 
        CHECK (mime_type IN ('image/jpeg', 'image/jpg', 'image/png', 'image/webp'));
    EXCEPTION WHEN OTHERS THEN
        NULL; -- Constraint already exists
    END;
    
    -- Add dimensions constraint if it doesn't exist
    BEGIN
        ALTER TABLE enhanced_product_images 
        ADD CONSTRAINT chk_enhanced_product_images_dimensions 
        CHECK (width <= 10000 AND height <= 10000);
    EXCEPTION WHEN OTHERS THEN
        NULL; -- Constraint already exists
    END;
    
    -- Add file_size constraint if it doesn't exist
    BEGIN
        ALTER TABLE enhanced_product_images 
        ADD CONSTRAINT chk_enhanced_product_images_file_size 
        CHECK (file_size <= 10485760); -- 10MB limit
    EXCEPTION WHEN OTHERS THEN
        NULL; -- Constraint already exists
    END;
END $$;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_enhanced_product_images_tenant_product ON enhanced_product_images (tenant_id, product_id);
CREATE INDEX IF NOT EXISTS idx_enhanced_product_images_tenant_created ON enhanced_product_images (tenant_id, created_at);
CREATE INDEX IF NOT EXISTS idx_enhanced_product_images_product_created ON enhanced_product_images (product_id, created_at);
CREATE INDEX IF NOT EXISTS idx_enhanced_product_images_mime_type ON enhanced_product_images (tenant_id, mime_type);
CREATE INDEX IF NOT EXISTS idx_enhanced_product_images_dimensions ON enhanced_product_images (tenant_id, width, height);

-- Create a function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_enhanced_product_images_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at (safely)
DO $$ BEGIN
    CREATE TRIGGER trigger_enhanced_product_images_updated_at
        BEFORE UPDATE ON enhanced_product_images
        FOR EACH ROW
        EXECUTE FUNCTION update_enhanced_product_images_updated_at();
EXCEPTION WHEN OTHERS THEN
    NULL; -- Trigger already exists
END $$;

-- Create a view for image statistics
CREATE OR REPLACE VIEW enhanced_product_image_stats AS
SELECT 
    tenant_id,
    product_id,
    COUNT(*) as image_count,
    SUM(file_size) as total_file_size,
    AVG(file_size) as avg_file_size,
    MAX(file_size) as max_file_size,
    MIN(file_size) as min_file_size,
    AVG(width) as avg_width,
    AVG(height) as avg_height,
    COUNT(DISTINCT mime_type) as mime_type_count,
    MIN(created_at) as first_image_date,
    MAX(created_at) as last_image_date
FROM enhanced_product_images
GROUP BY tenant_id, product_id;

-- Create a function to migrate existing product images to enhanced format
CREATE OR REPLACE FUNCTION migrate_product_images_to_enhanced()
RETURNS INTEGER AS $$
DECLARE
    migrated_count INTEGER := 0;
    image_record RECORD;
BEGIN
    -- Migrate existing product images to enhanced format
    FOR image_record IN 
        SELECT id, tenant_id, product_id, image_url, alt_text, created_at
        FROM product_images
        WHERE id NOT IN (
            SELECT id FROM enhanced_product_images 
            WHERE enhanced_product_images.id = product_images.id
        )
    LOOP
        INSERT INTO enhanced_product_images (
            id, tenant_id, product_id, original_url, alt_text, 
            file_size, mime_type, width, height, created_at, updated_at
        ) VALUES (
            image_record.id,
            image_record.tenant_id,
            image_record.product_id,
            image_record.image_url,
            image_record.alt_text,
            0, -- Default file size (unknown for existing images)
            'image/jpeg', -- Default MIME type
            800, -- Default width
            600, -- Default height
            image_record.created_at,
            NOW()
        );
        
        migrated_count := migrated_count + 1;
    END LOOP;
    
    RETURN migrated_count;
END;
$$ LANGUAGE plpgsql;

-- Create a function to cleanup orphaned image files
CREATE OR REPLACE FUNCTION cleanup_orphaned_enhanced_images()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete enhanced images where the product no longer exists
    DELETE FROM enhanced_product_images 
    WHERE product_id NOT IN (SELECT id FROM products);
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON TABLE enhanced_product_images IS 'Enhanced product images with multiple size variants and metadata';
COMMENT ON COLUMN enhanced_product_images.original_url IS 'URL to the original uploaded image';
COMMENT ON COLUMN enhanced_product_images.large_url IS 'URL to the large size variant (1920px)';
COMMENT ON COLUMN enhanced_product_images.medium_url IS 'URL to the medium size variant (800px)';
COMMENT ON COLUMN enhanced_product_images.small_url IS 'URL to the small size variant (300px)';
COMMENT ON COLUMN enhanced_product_images.thumbnail_url IS 'URL to the thumbnail variant (150px)';
COMMENT ON COLUMN enhanced_product_images.file_size IS 'Original file size in bytes';
COMMENT ON COLUMN enhanced_product_images.mime_type IS 'MIME type of the original image';
COMMENT ON COLUMN enhanced_product_images.width IS 'Original image width in pixels';
COMMENT ON COLUMN enhanced_product_images.height IS 'Original image height in pixels';

COMMENT ON FUNCTION migrate_product_images_to_enhanced IS 'Migrates existing product images to the enhanced format';
COMMENT ON FUNCTION cleanup_orphaned_enhanced_images IS 'Removes enhanced images for products that no longer exist';