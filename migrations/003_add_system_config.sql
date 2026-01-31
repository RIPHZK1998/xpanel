-- Migration: Add system_config table for hot-reloadable configuration
-- This allows updating system settings (like Node API key) without server restart

CREATE TABLE IF NOT EXISTS system_config (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    encrypted BOOLEAN DEFAULT FALSE,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by INTEGER REFERENCES users(id) ON DELETE SET NULL
);

-- Create index for faster lookups
CREATE INDEX idx_system_config_updated_at ON system_config(updated_at DESC);

-- Seed default configuration
-- Note: NODE_API_KEY should be set via admin panel after first run
INSERT INTO system_config (key, value, encrypted, description) VALUES
    ('node_api_key', '', TRUE, 'API key used by nodes to authenticate with the panel')
ON CONFLICT (key) DO NOTHING;

-- Add comment
COMMENT ON TABLE system_config IS 'System-wide configuration with encryption support for sensitive values';
