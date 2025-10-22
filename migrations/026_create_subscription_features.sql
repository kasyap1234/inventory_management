-- Migration 026: Create Subscription Features and Limits
-- This migration creates tables for managing subscription plan features and usage tracking

-- Create subscription_features table to store plan-based feature limits
CREATE TABLE IF NOT EXISTS subscription_features (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id VARCHAR(50) NOT NULL UNIQUE,
    plan_name VARCHAR(100) NOT NULL,
    
    -- Feature Limits
    max_warehouses INT NOT NULL DEFAULT 5,
    max_users INT NOT NULL DEFAULT 5,
    max_products INT NOT NULL DEFAULT 1000,
    max_orders_per_month INT NOT NULL DEFAULT 100,
    max_suppliers INT NOT NULL DEFAULT 50,
    max_distributors INT NOT NULL DEFAULT 50,
    
    -- Feature Access Flags
    analytics_enabled BOOLEAN DEFAULT FALSE,
    advanced_analytics_enabled BOOLEAN DEFAULT FALSE,
    api_access_enabled BOOLEAN DEFAULT FALSE,
    custom_branding_enabled BOOLEAN DEFAULT FALSE,
    priority_support_enabled BOOLEAN DEFAULT FALSE,
    multi_location_enabled BOOLEAN DEFAULT FALSE,
    custom_integrations_enabled BOOLEAN DEFAULT FALSE,
    dedicated_account_manager BOOLEAN DEFAULT FALSE,
    
    -- Limits
    api_rate_limit_per_minute INT DEFAULT 60,
    storage_limit_gb INT DEFAULT 10,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Insert default plan features
INSERT INTO subscription_features (
    plan_id, plan_name, 
    max_warehouses, max_users, max_products, max_orders_per_month, 
    max_suppliers, max_distributors,
    analytics_enabled, advanced_analytics_enabled, api_access_enabled,
    custom_branding_enabled, priority_support_enabled, multi_location_enabled,
    custom_integrations_enabled, dedicated_account_manager,
    api_rate_limit_per_minute, storage_limit_gb
) VALUES
-- Basic Plan (Starter tier)
('basic', 'Basic Plan',
    5, 5, 1000, 100,
    50, 50,
    TRUE, FALSE, FALSE,
    FALSE, FALSE, FALSE,
    FALSE, FALSE,
    60, 10
),
-- Premium Plan (Growth tier)
('premium', 'Premium Plan',
    20, 20, 10000, 1000,
    200, 200,
    TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE,
    FALSE, FALSE,
    300, 100
),
-- Enterprise Plan (Scale tier)
('enterprise', 'Enterprise Plan',
    -1, -1, -1, -1,  -- -1 means unlimited
    -1, -1,
    TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE,
    TRUE, TRUE,
    1000, 1000
)
ON CONFLICT (plan_id) DO UPDATE SET
    max_warehouses = EXCLUDED.max_warehouses,
    max_users = EXCLUDED.max_users,
    max_products = EXCLUDED.max_products,
    max_orders_per_month = EXCLUDED.max_orders_per_month,
    max_suppliers = EXCLUDED.max_suppliers,
    max_distributors = EXCLUDED.max_distributors,
    analytics_enabled = EXCLUDED.analytics_enabled,
    advanced_analytics_enabled = EXCLUDED.advanced_analytics_enabled,
    api_access_enabled = EXCLUDED.api_access_enabled,
    custom_branding_enabled = EXCLUDED.custom_branding_enabled,
    priority_support_enabled = EXCLUDED.priority_support_enabled,
    multi_location_enabled = EXCLUDED.multi_location_enabled,
    custom_integrations_enabled = EXCLUDED.custom_integrations_enabled,
    dedicated_account_manager = EXCLUDED.dedicated_account_manager,
    api_rate_limit_per_minute = EXCLUDED.api_rate_limit_per_minute,
    storage_limit_gb = EXCLUDED.storage_limit_gb,
    updated_at = NOW();

-- Create usage_tracking table for tracking feature usage
CREATE TABLE IF NOT EXISTS usage_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
    
    -- Current usage counts
    warehouses_count INT DEFAULT 0,
    users_count INT DEFAULT 0,
    products_count INT DEFAULT 0,
    orders_count_current_month INT DEFAULT 0,
    suppliers_count INT DEFAULT 0,
    distributors_count INT DEFAULT 0,
    
    -- Storage tracking
    storage_used_gb DECIMAL(10,2) DEFAULT 0,
    
    -- API usage
    api_calls_current_month INT DEFAULT 0,
    
    -- Tracking period
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE (tenant_id, period_start)
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_usage_tracking_tenant ON usage_tracking(tenant_id);
CREATE INDEX IF NOT EXISTS idx_usage_tracking_subscription ON usage_tracking(subscription_id);
CREATE INDEX IF NOT EXISTS idx_usage_tracking_period ON usage_tracking(period_start, period_end);

-- Add function to get current usage for a tenant
CREATE OR REPLACE FUNCTION get_current_usage(p_tenant_id UUID)
RETURNS TABLE (
    warehouses_count INT,
    users_count INT,
    products_count INT,
    orders_count_current_month INT,
    suppliers_count INT,
    distributors_count INT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        (SELECT COUNT(*)::INT FROM warehouses WHERE tenant_id = p_tenant_id AND deleted_at IS NULL),
        (SELECT COUNT(*)::INT FROM users WHERE tenant_id = p_tenant_id),
        (SELECT COUNT(*)::INT FROM products WHERE tenant_id = p_tenant_id AND deleted_at IS NULL),
        (SELECT COUNT(*)::INT FROM orders WHERE tenant_id = p_tenant_id 
         AND created_at >= date_trunc('month', CURRENT_DATE)),
        (SELECT COUNT(*)::INT FROM suppliers WHERE tenant_id = p_tenant_id AND deleted_at IS NULL),
        (SELECT COUNT(*)::INT FROM distributors WHERE tenant_id = p_tenant_id AND deleted_at IS NULL);
END;
$$ LANGUAGE plpgsql;

-- Add function to check if tenant can access feature
CREATE OR REPLACE FUNCTION can_access_feature(
    p_tenant_id UUID,
    p_feature_name VARCHAR
)
RETURNS BOOLEAN AS $$
DECLARE
    v_plan_id VARCHAR(50);
    v_result BOOLEAN;
BEGIN
    -- Get active subscription plan for tenant
    SELECT 
        CASE 
            WHEN s.plan_name ILIKE '%basic%' THEN 'basic'
            WHEN s.plan_name ILIKE '%premium%' THEN 'premium'
            WHEN s.plan_name ILIKE '%enterprise%' THEN 'enterprise'
            ELSE 'basic'
        END INTO v_plan_id
    FROM subscriptions s
    WHERE s.tenant_id = p_tenant_id 
        AND s.status = 'active'
    ORDER BY s.created_at DESC
    LIMIT 1;
    
    -- If no subscription, return false
    IF v_plan_id IS NULL THEN
        RETURN FALSE;
    END IF;
    
    -- Check feature access
    SELECT 
        CASE p_feature_name
            WHEN 'analytics' THEN sf.analytics_enabled
            WHEN 'advanced_analytics' THEN sf.advanced_analytics_enabled
            WHEN 'api_access' THEN sf.api_access_enabled
            WHEN 'custom_branding' THEN sf.custom_branding_enabled
            WHEN 'priority_support' THEN sf.priority_support_enabled
            WHEN 'multi_location' THEN sf.multi_location_enabled
            WHEN 'custom_integrations' THEN sf.custom_integrations_enabled
            WHEN 'dedicated_account_manager' THEN sf.dedicated_account_manager
            ELSE FALSE
        END INTO v_result
    FROM subscription_features sf
    WHERE sf.plan_id = v_plan_id;
    
    RETURN COALESCE(v_result, FALSE);
END;
$$ LANGUAGE plpgsql;

-- Create trigger to update usage_tracking automatically
CREATE OR REPLACE FUNCTION update_usage_tracking()
RETURNS TRIGGER AS $$
DECLARE
    v_tenant_id UUID;
    v_current_period_start TIMESTAMP;
    v_current_period_end TIMESTAMP;
BEGIN
    -- Get tenant_id from the operation
    v_tenant_id := COALESCE(NEW.tenant_id, OLD.tenant_id);
    
    -- Get current period (monthly)
    v_current_period_start := date_trunc('month', CURRENT_DATE);
    v_current_period_end := (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month' - INTERVAL '1 day');
    
    -- Update or insert usage tracking
    INSERT INTO usage_tracking (
        tenant_id, period_start, period_end,
        warehouses_count, users_count, products_count,
        orders_count_current_month, suppliers_count, distributors_count
    )
    SELECT 
        v_tenant_id, v_current_period_start, v_current_period_end,
        u.warehouses_count, u.users_count, u.products_count,
        u.orders_count_current_month, u.suppliers_count, u.distributors_count
    FROM get_current_usage(v_tenant_id) u
    ON CONFLICT (tenant_id, period_start) DO UPDATE SET
        warehouses_count = EXCLUDED.warehouses_count,
        users_count = EXCLUDED.users_count,
        products_count = EXCLUDED.products_count,
        orders_count_current_month = EXCLUDED.orders_count_current_month,
        suppliers_count = EXCLUDED.suppliers_count,
        distributors_count = EXCLUDED.distributors_count,
        updated_at = NOW();
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach triggers to relevant tables (optional - for automatic tracking)
-- Note: For better performance, you may want to update usage through application code instead
-- CREATE TRIGGER trg_update_usage_warehouses
--     AFTER INSERT OR UPDATE OR DELETE ON warehouses
--     FOR EACH ROW EXECUTE FUNCTION update_usage_tracking();

COMMENT ON TABLE subscription_features IS 'Stores feature limits and capabilities for each subscription plan';
COMMENT ON TABLE usage_tracking IS 'Tracks resource usage per tenant for subscription limit enforcement';
COMMENT ON FUNCTION get_current_usage IS 'Returns current usage counts for a tenant';
COMMENT ON FUNCTION can_access_feature IS 'Checks if tenant can access a specific feature based on their subscription';
