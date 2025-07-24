-- Secure Provider Credentials Migration
-- This migration creates a new table for secure credential management
-- API keys are stored in environment variables, not in the database

-- Create new secure provider credentials table
CREATE TABLE IF NOT EXISTS secure_provider_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,                        -- Provider type (godaddy, namecheap, hostinger)
    name VARCHAR(100) NOT NULL,                           -- User-friendly name
    account_name VARCHAR(255) NOT NULL,                   -- Account identifier (email/username)
    credential_reference VARCHAR(100) NOT NULL,           -- Reference to environment variables (e.g., "GODADDY_DEFAULT")
    enabled BOOLEAN NOT NULL DEFAULT true,
    connection_status VARCHAR(50) NOT NULL DEFAULT 'disconnected', -- connected, error, testing, disconnected
    last_sync TIMESTAMPTZ,
    last_sync_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name, provider),
    UNIQUE(credential_reference, provider)                -- Prevent duplicate references per provider
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_secure_credentials_provider ON secure_provider_credentials(provider);
CREATE INDEX IF NOT EXISTS idx_secure_credentials_status ON secure_provider_credentials(connection_status);
CREATE INDEX IF NOT EXISTS idx_secure_credentials_reference ON secure_provider_credentials(credential_reference);
CREATE INDEX IF NOT EXISTS idx_secure_credentials_enabled ON secure_provider_credentials(enabled);

-- Insert default secure credential references if environment variables exist
-- Note: These inserts will only work if the corresponding environment variables are set

-- Example default connections (uncomment and modify as needed)
-- INSERT INTO secure_provider_credentials (provider, name, account_name, credential_reference, enabled)
-- SELECT 'godaddy', 'GoDaddy Production', 'production-account', 'GODADDY_DEFAULT', true
-- WHERE EXISTS (SELECT 1 FROM pg_settings WHERE name = 'GODADDY_API_KEY' AND setting != '');

-- INSERT INTO secure_provider_credentials (provider, name, account_name, credential_reference, enabled)
-- SELECT 'namecheap', 'Namecheap Production', 'production-account', 'NAMECHEAP_DEFAULT', true
-- WHERE EXISTS (SELECT 1 FROM pg_settings WHERE name = 'NAMECHEAP_API_KEY' AND setting != '');

-- INSERT INTO secure_provider_credentials (provider, name, account_name, credential_reference, enabled)
-- SELECT 'hostinger', 'Hostinger Production', 'production-account', 'HOSTINGER_DEFAULT', true
-- WHERE EXISTS (SELECT 1 FROM pg_settings WHERE name = 'HOSTINGER_API_KEY' AND setting != '');

-- Add a comment to document the security improvement
COMMENT ON TABLE secure_provider_credentials IS 'Secure provider credentials table. API keys are stored in environment variables, not in the database. The credential_reference field points to predefined environment variable sets.';
COMMENT ON COLUMN secure_provider_credentials.credential_reference IS 'Reference to environment variable set (e.g., GODADDY_DEFAULT, NAMECHEAP_STAGING). Actual API keys are loaded from .env file.';

-- Future migration path: Gradually migrate from old provider_credentials table
-- Step 1: Create secure_provider_credentials table (this migration)
-- Step 2: Migrate existing connections to use environment variable references
-- Step 3: Deprecate and eventually drop the old provider_credentials table

-- Security Notes:
-- 1. API keys are never stored in the database
-- 2. Only references to environment variable sets are stored
-- 3. Environment variables should be managed securely (.env file, secrets manager, etc.)
-- 4. Multiple environments (staging, production) are supported through different references
-- 5. Credential rotation only requires updating environment variables, not database records
