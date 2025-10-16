-- Enhanced Notification System Tables
-- This migration adds comprehensive notification system support including templates, webhooks, and alerts

-- Notification Templates
CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('email', 'sms', 'webhook', 'push')),
    event_type VARCHAR(100) NOT NULL, -- e.g., 'order_created', 'low_stock', 'payment_received'
    subject VARCHAR(500), -- For email templates
    body_template TEXT NOT NULL,
    variables JSONB DEFAULT '{}', -- Template variables and their descriptions
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (tenant_id, name, type)
);

-- Webhook Subscriptions
CREATE TABLE IF NOT EXISTS webhook_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(1000) NOT NULL,
    secret VARCHAR(255), -- For webhook signature verification
    event_types TEXT[] NOT NULL, -- Array of event types to subscribe to
    headers JSONB DEFAULT '{}', -- Custom headers to send with webhook
    timeout_seconds INTEGER DEFAULT 30 CHECK (timeout_seconds > 0 AND timeout_seconds <= 300),
    retry_count INTEGER DEFAULT 3 CHECK (retry_count >= 0 AND retry_count <= 10),
    is_active BOOLEAN DEFAULT true,
    last_success_at TIMESTAMP,
    last_failure_at TIMESTAMP,
    failure_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

-- Alert Rules
CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    conditions JSONB NOT NULL, -- Rule conditions (e.g., {"field": "quantity", "operator": "lt", "value": 10})
    actions JSONB NOT NULL, -- Actions to take when rule matches
    is_active BOOLEAN DEFAULT true,
    last_triggered_at TIMESTAMP,
    trigger_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

-- Notification Configs (User preferences for notifications)
CREATE TABLE IF NOT EXISTS notification_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    channels TEXT[] NOT NULL DEFAULT '{}', -- Array of preferred channels: email, sms, push, webhook
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (tenant_id, user_id, event_type)
);

-- Notification Delivery Log (Track delivery attempts and status)
CREATE TABLE IF NOT EXISTS notification_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    notification_id UUID REFERENCES notifications(id) ON DELETE CASCADE,
    template_id UUID REFERENCES notification_templates(id) ON DELETE SET NULL,
    webhook_id UUID REFERENCES webhook_subscriptions(id) ON DELETE SET NULL,
    channel VARCHAR(50) NOT NULL CHECK (channel IN ('email', 'sms', 'webhook', 'push')),
    recipient VARCHAR(500) NOT NULL, -- Email, phone number, webhook URL, etc.
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'delivered', 'failed', 'bounced')),
    attempt_count INTEGER DEFAULT 0,
    last_attempt_at TIMESTAMP,
    delivered_at TIMESTAMP,
    error_message TEXT,
    response_data JSONB, -- Store webhook responses, delivery receipts, etc.
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_notification_templates_tenant_type ON notification_templates (tenant_id, type);
CREATE INDEX IF NOT EXISTS idx_notification_templates_tenant_event ON notification_templates (tenant_id, event_type);
CREATE INDEX IF NOT EXISTS idx_notification_templates_active ON notification_templates (tenant_id, is_active);

CREATE INDEX IF NOT EXISTS idx_webhook_subscriptions_tenant_active ON webhook_subscriptions (tenant_id, is_active);
CREATE INDEX IF NOT EXISTS idx_webhook_subscriptions_event_types ON webhook_subscriptions USING GIN (event_types);

CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant_event ON alert_rules (tenant_id, event_type);
CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant_active ON alert_rules (tenant_id, is_active);

CREATE INDEX IF NOT EXISTS idx_notification_configs_tenant_user ON notification_configs (tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_notification_configs_event_enabled ON notification_configs (tenant_id, event_type, is_enabled);

CREATE INDEX IF NOT EXISTS idx_notification_deliveries_tenant_status ON notification_deliveries (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_notification_deliveries_notification ON notification_deliveries (notification_id);
CREATE INDEX IF NOT EXISTS idx_notification_deliveries_created ON notification_deliveries (tenant_id, created_at);
CREATE INDEX IF NOT EXISTS idx_notification_deliveries_channel ON notification_deliveries (tenant_id, channel, status);