-- Initial schema
CREATE TABLE domains (
    id UUID PRIMARY KEY,
    name VARCHAR(253) NOT NULL UNIQUE,
    provider VARCHAR(50) NOT NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    category_id UUID,
    project_id UUID,
    auto_renew BOOLEAN NOT NULL DEFAULT true,
    renewal_price NUMERIC(10,2),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    tags JSONB DEFAULT '[]'
);

-- Categories table
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    color VARCHAR(7) DEFAULT '#6366f1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Projects table
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    color VARCHAR(7) DEFAULT '#059669',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Provider credentials table
CREATE TABLE provider_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    credentials JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_sync TIMESTAMPTZ,
    last_sync_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name, provider)
);

-- Users table for admin authentication
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'admin',
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Sessions table for authentication
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- DNS records table
CREATE TABLE dns_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain_id UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL,
    name VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    ttl INTEGER NOT NULL DEFAULT 3600,
    priority INTEGER,
    weight INTEGER,
    port INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_domain_expiry ON domains(expires_at);
CREATE INDEX idx_domain_category ON domains(category_id);
CREATE INDEX idx_domain_project ON domains(project_id);
CREATE INDEX idx_domain_status ON domains(status);
CREATE INDEX idx_credentials_provider ON provider_credentials(provider);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
CREATE INDEX idx_dns_domain ON dns_records(domain_id);
CREATE INDEX idx_dns_type ON dns_records(type);

-- Foreign key constraints
ALTER TABLE domains ADD CONSTRAINT fk_domain_category 
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
ALTER TABLE domains ADD CONSTRAINT fk_domain_project 
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;

-- Insert default categories
INSERT INTO categories (name, description, color) VALUES 
    ('Personal', 'Personal domains', '#6366f1'),
    ('Business', 'Business domains', '#dc2626'),
    ('Development', 'Development and testing domains', '#059669'),
    ('Client Work', 'Client project domains', '#d97706'),
    ('Investment', 'Domain investments', '#7c3aed');

-- Insert default projects
INSERT INTO projects (name, description, color) VALUES 
    ('Portfolio Sites', 'Personal portfolio websites', '#059669'),
    ('E-commerce', 'Online store projects', '#dc2626'),
    ('SaaS Products', 'Software as a Service projects', '#7c3aed'),
    ('Marketing Campaigns', 'Marketing and landing pages', '#d97706');

-- Insert default admin user (password: admin123 - CHANGE THIS IN PRODUCTION!)
INSERT INTO users (username, email, password_hash, role) VALUES 
    ('admin', 'admin@domainvault.local', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj0GdRLXu6G6', 'admin');

-- Future scaling
-- ALTER TABLE domains ADD COLUMN renewal_price NUMERIC;
-- ALTER TABLE domains ADD COLUMN dns_records JSONB;
