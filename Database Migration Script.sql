-- Initial schema
CREATE TABLE domains (
    id UUID PRIMARY KEY,
    name VARCHAR(253) NOT NULL UNIQUE,
    provider VARCHAR(50) NOT NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Future alerting index
CREATE INDEX idx_domain_expiry ON domains(expires_at);

-- Future scaling
-- ALTER TABLE domains ADD COLUMN renewal_price NUMERIC;
-- ALTER TABLE domains ADD COLUMN dns_records JSONB;