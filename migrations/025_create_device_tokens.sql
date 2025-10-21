-- Migration: Create device tokens table for push notifications
-- Description: Adds support for FCM/APNs device token management
-- Author: System
-- Date: 2025-10-21

-- Create device_tokens table
CREATE TABLE IF NOT EXISTS device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token VARCHAR(500) NOT NULL,
    device_type VARCHAR(50) NOT NULL CHECK (device_type IN ('android', 'ios', 'web')),
    device_name VARCHAR(255),
    app_version VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Ensure unique tokens per tenant and user
    CONSTRAINT unique_device_token UNIQUE (tenant_id, user_id, device_token)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_device_tokens_tenant_user ON device_tokens(tenant_id, user_id, is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_device_tokens_token ON device_tokens(device_token) WHERE is_active = true AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_device_tokens_user ON device_tokens(user_id) WHERE is_active = true AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_device_tokens_type ON device_tokens(device_type) WHERE is_active = true AND deleted_at IS NULL;

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_device_tokens_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_device_tokens_updated_at
    BEFORE UPDATE ON device_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_device_tokens_updated_at();

-- Add comments for documentation
COMMENT ON TABLE device_tokens IS 'Stores FCM/APNs device tokens for push notifications';
COMMENT ON COLUMN device_tokens.device_token IS 'FCM or APNs device registration token';
COMMENT ON COLUMN device_tokens.device_type IS 'Device platform: android, ios, or web';
COMMENT ON COLUMN device_tokens.is_active IS 'Whether the token is still valid and active';
COMMENT ON COLUMN device_tokens.last_used_at IS 'Last time a notification was successfully sent to this token';
