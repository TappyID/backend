-- Migration: Create user_connections table
-- Description: Store user connection status for social platforms (WhatsApp, etc.)

CREATE TABLE IF NOT EXISTS user_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    platform VARCHAR(50) NOT NULL, -- 'whatsapp', 'facebook', 'instagram', etc.
    status VARCHAR(20) NOT NULL DEFAULT 'disconnected', -- 'connected', 'disconnected', 'connecting', 'error'
    session_name VARCHAR(100), -- Session identifier for the platform
    session_data JSONB, -- Additional session metadata
    connected_at TIMESTAMP,
    disconnected_at TIMESTAMP,
    last_sync_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure one connection per user per platform
    UNIQUE(user_id, platform)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_user_connections_user_id ON user_connections(user_id);
CREATE INDEX IF NOT EXISTS idx_user_connections_platform ON user_connections(platform);
CREATE INDEX IF NOT EXISTS idx_user_connections_status ON user_connections(status);

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION update_user_connections_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_connections_updated_at
    BEFORE UPDATE ON user_connections
    FOR EACH ROW
    EXECUTE FUNCTION update_user_connections_updated_at();
