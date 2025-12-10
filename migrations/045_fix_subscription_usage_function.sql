-- Migration 045: Fix get_current_usage function (remove deleted_at dependency)
-- The current tables (warehouses, products, suppliers, distributors) do not have a deleted_at column.
-- The previous version of get_current_usage referenced deleted_at, causing runtime errors (SQLSTATE 42703).

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
        (SELECT COUNT(*)::INT FROM warehouses  WHERE tenant_id = p_tenant_id),
        (SELECT COUNT(*)::INT FROM users       WHERE tenant_id = p_tenant_id),
        (SELECT COUNT(*)::INT FROM products    WHERE tenant_id = p_tenant_id),
        (SELECT COUNT(*)::INT FROM orders      WHERE tenant_id = p_tenant_id 
         AND created_at >= date_trunc('month', CURRENT_DATE)),
        (SELECT COUNT(*)::INT FROM suppliers   WHERE tenant_id = p_tenant_id),
        (SELECT COUNT(*)::INT FROM distributors WHERE tenant_id = p_tenant_id);
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_current_usage IS 'Returns current usage counts for a tenant (warehouses, users, products, orders this month, suppliers, distributors)';

