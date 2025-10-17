-- Enhance Audit Logs Table
-- This migration enhances the existing audit_logs table to support comprehensive audit logging

-- Drop dependent views first
DROP VIEW IF EXISTS audit_log_summary CASCADE;

-- Add new columns to audit_logs table
ALTER TABLE audit_logs 
ADD COLUMN IF NOT EXISTS request_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS ip_address INET,
ADD COLUMN IF NOT EXISTS user_agent TEXT,
ADD COLUMN IF NOT EXISTS method VARCHAR(10),
ADD COLUMN IF NOT EXISTS path VARCHAR(500),
ADD COLUMN IF NOT EXISTS status_code INTEGER,
ADD COLUMN IF NOT EXISTS success BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS error_message TEXT,
ADD COLUMN IF NOT EXISTS old_values JSONB,
ADD COLUMN IF NOT EXISTS new_values JSONB,
ADD COLUMN IF NOT EXISTS changes JSONB,
ADD COLUMN IF NOT EXISTS metadata JSONB;

-- Alter column types (safe now that views are dropped)
DO $$ BEGIN
    BEGIN
        ALTER TABLE audit_logs ALTER COLUMN action TYPE VARCHAR(50);
    EXCEPTION WHEN OTHERS THEN
        NULL;
    END;
    BEGIN
        ALTER TABLE audit_logs ALTER COLUMN table_name TYPE VARCHAR(100);
    EXCEPTION WHEN OTHERS THEN
        NULL;
    END;
END $$;

-- Add constraints (using plain SQL without DO blocks)
-- Note: This migration assumes the constraints might already exist, so we use DO blocks instead

-- Create additional indexes for enhanced audit logging
CREATE INDEX IF NOT EXISTS idx_audit_logs_request_id ON audit_logs (tenant_id, request_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action_resource ON audit_logs (tenant_id, action, table_name);
CREATE INDEX IF NOT EXISTS idx_audit_logs_success ON audit_logs (tenant_id, success);
CREATE INDEX IF NOT EXISTS idx_audit_logs_ip_address ON audit_logs (tenant_id, ip_address);
CREATE INDEX IF NOT EXISTS idx_audit_logs_method_path ON audit_logs (tenant_id, method, path);

-- Create composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_action_date ON audit_logs (tenant_id, changed_by, action, created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_date ON audit_logs (tenant_id, table_name, created_at);

-- Create partial indexes for performance (removed NOW() function calls in indexes as they're not allowed)
CREATE INDEX IF NOT EXISTS idx_audit_logs_failures ON audit_logs (tenant_id, created_at) WHERE success = false;

-- Add GIN indexes for JSONB columns
CREATE INDEX IF NOT EXISTS idx_audit_logs_old_values_gin ON audit_logs USING GIN (old_values);
CREATE INDEX IF NOT EXISTS idx_audit_logs_new_values_gin ON audit_logs USING GIN (new_values);
CREATE INDEX IF NOT EXISTS idx_audit_logs_changes_gin ON audit_logs USING GIN (changes);
CREATE INDEX IF NOT EXISTS idx_audit_logs_metadata_gin ON audit_logs USING GIN (metadata);

-- Create a view for audit log summaries
CREATE OR REPLACE VIEW audit_log_summary AS
SELECT 
    tenant_id,
    action,
    table_name as resource,
    DATE_TRUNC('day', created_at) as log_date,
    COUNT(*) as total_count,
    COUNT(*) FILTER (WHERE success = true) as success_count,
    COUNT(*) FILTER (WHERE success = false) as failure_count,
    COUNT(DISTINCT changed_by) as unique_users,
    MIN(created_at) as first_occurrence,
    MAX(created_at) as last_occurrence
FROM audit_logs
GROUP BY tenant_id, action, table_name, DATE_TRUNC('day', created_at);

-- Create a function to clean up old audit logs
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs(retention_days INTEGER DEFAULT 365)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM audit_logs 
    WHERE created_at < NOW() - (retention_days || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Log the cleanup operation
    INSERT INTO audit_logs (
        tenant_id, action, table_name, record_id, 
        success, metadata, created_at
    ) VALUES (
        '00000000-0000-0000-0000-000000000000'::UUID, 
        'cleanup', 
        'audit_logs', 
        'system',
        true,
        jsonb_build_object(
            'deleted_count', deleted_count,
            'retention_days', retention_days,
            'cleanup_date', NOW()
        ),
        NOW()
    );
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create a function to get audit statistics
CREATE OR REPLACE FUNCTION get_audit_statistics(
    p_tenant_id UUID,
    p_start_date TIMESTAMP DEFAULT NOW() - INTERVAL '30 days',
    p_end_date TIMESTAMP DEFAULT NOW()
)
RETURNS TABLE (
    action VARCHAR(50),
    resource VARCHAR(100),
    total_count BIGINT,
    success_count BIGINT,
    failure_count BIGINT,
    success_rate NUMERIC(5,2)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        al.action,
        al.table_name as resource,
        COUNT(*) as total_count,
        COUNT(*) FILTER (WHERE al.success = true) as success_count,
        COUNT(*) FILTER (WHERE al.success = false) as failure_count,
        ROUND(
            (COUNT(*) FILTER (WHERE al.success = true) * 100.0 / COUNT(*)), 
            2
        ) as success_rate
    FROM audit_logs al
    WHERE al.tenant_id = p_tenant_id
      AND al.created_at BETWEEN p_start_date AND p_end_date
    GROUP BY al.action, al.table_name
    ORDER BY total_count DESC;
END;
$$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON TABLE audit_logs IS 'Comprehensive audit log table tracking all system operations';
COMMENT ON COLUMN audit_logs.request_id IS 'Unique identifier for the HTTP request that triggered this audit event';
COMMENT ON COLUMN audit_logs.old_values IS 'JSON representation of the record state before the operation';
COMMENT ON COLUMN audit_logs.new_values IS 'JSON representation of the record state after the operation';
COMMENT ON COLUMN audit_logs.changes IS 'JSON representation of the specific changes made';
COMMENT ON COLUMN audit_logs.metadata IS 'Additional contextual information about the operation';
COMMENT ON FUNCTION cleanup_old_audit_logs IS 'Removes audit logs older than the specified retention period';
COMMENT ON FUNCTION get_audit_statistics IS 'Returns audit statistics for a tenant within a date range';
