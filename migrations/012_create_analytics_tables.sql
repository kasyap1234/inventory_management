-- Migration: Create analytics tables for search and bulk operation tracking
-- Created: 2025-09-01

-- Create search_analytics table for tracking search operations
CREATE TABLE IF NOT EXISTS search_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL, -- 'products', 'orders', 'inventory', etc.
    search_term VARCHAR(255),
    filter_count INTEGER DEFAULT 0,
    result_count INTEGER DEFAULT 0,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    response_time_ms BIGINT DEFAULT 0,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for search_analytics
CREATE INDEX IF NOT EXISTS idx_search_analytics_tenant_id ON search_analytics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_search_analytics_entity_type ON search_analytics(entity_type);
CREATE INDEX IF NOT EXISTS idx_search_analytics_timestamp ON search_analytics(timestamp);
CREATE INDEX IF NOT EXISTS idx_search_analytics_search_term ON search_analytics(search_term);

-- Create bulk_operation_analytics table for tracking bulk operations
CREATE TABLE IF NOT EXISTS bulk_operation_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    operation_type VARCHAR(100) NOT NULL, -- 'product_bulk_update', 'inventory_bulk_adjust', etc.
    total_items INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    failure_count INTEGER NOT NULL DEFAULT 0,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    processing_time_ms BIGINT DEFAULT 0,
    error_summary TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for bulk_operation_analytics
CREATE INDEX IF NOT EXISTS idx_bulk_operation_analytics_tenant_id ON bulk_operation_analytics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_bulk_operation_analytics_operation_type ON bulk_operation_analytics(operation_type);
CREATE INDEX IF NOT EXISTS idx_bulk_operation_analytics_timestamp ON bulk_operation_analytics(timestamp);

-- Create tenant_analytics_cache table for caching aggregated analytics
CREATE TABLE IF NOT EXISTS tenant_analytics_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    metric_type VARCHAR(100) NOT NULL, -- 'sales', 'inventory', 'search_stats', etc.
    metric_data JSONB NOT NULL,
    calculated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, metric_type)
);

-- Create indexes for tenant_analytics_cache
CREATE INDEX IF NOT EXISTS idx_tenant_analytics_cache_tenant_id ON tenant_analytics_cache(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_analytics_cache_metric_type ON tenant_analytics_cache(metric_type);
CREATE INDEX IF NOT EXISTS idx_tenant_analytics_cache_expires_at ON tenant_analytics_cache(expires_at);

-- Add comments for documentation
COMMENT ON TABLE search_analytics IS 'Tracks all search operations across the application for analytics';
COMMENT ON TABLE bulk_operation_analytics IS 'Tracks bulk operation performance and results';
COMMENT ON TABLE tenant_analytics_cache IS 'Caches aggregated analytics data to improve query performance';
