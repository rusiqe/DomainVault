-- Enhanced Provider Credentials Migration
-- Add new fields to support better provider management

-- Add new columns to provider_credentials table
ALTER TABLE provider_credentials 
ADD COLUMN IF NOT EXISTS account_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS connection_status VARCHAR(50) DEFAULT 'disconnected';

-- Update existing records with default values
UPDATE provider_credentials 
SET account_name = name WHERE account_name IS NULL;

UPDATE provider_credentials 
SET connection_status = 'connected' WHERE enabled = true;

UPDATE provider_credentials 
SET connection_status = 'disconnected' WHERE enabled = false;

-- Create index on connection_status for better query performance
CREATE INDEX IF NOT EXISTS idx_provider_credentials_status ON provider_credentials(connection_status);
CREATE INDEX IF NOT EXISTS idx_provider_credentials_account ON provider_credentials(account_name);
