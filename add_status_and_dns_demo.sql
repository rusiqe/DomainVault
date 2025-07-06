-- Add status column to domains table for HTTP status tracking
ALTER TABLE domains ADD COLUMN IF NOT EXISTS http_status INTEGER;
ALTER TABLE domains ADD COLUMN IF NOT EXISTS last_status_check TIMESTAMPTZ;
ALTER TABLE domains ADD COLUMN IF NOT EXISTS status_message TEXT;

-- Update existing domains with realistic HTTP status
UPDATE domains SET 
    http_status = CASE 
        WHEN status = 'expired' THEN 404
        WHEN name LIKE '%tech%' OR name LIKE '%solutions%' THEN 200
        WHEN name LIKE '%photography%' OR name LIKE '%boutique%' THEN 200
        WHEN name LIKE '%dev%' OR name LIKE '%staging%' OR name LIKE '%testing%' THEN 503
        WHEN name LIKE '%urgent%' OR name LIKE '%almost%' THEN 200
        WHEN name LIKE '%premium%' OR name LIKE '%future%' THEN 301
        WHEN name LIKE '%blockchain%' THEN 403
        WHEN name LIKE '%sale%' OR name LIKE '%launch%' THEN 200
        ELSE 200
    END,
    last_status_check = NOW() - (RANDOM() * INTERVAL '24 hours'),
    status_message = CASE 
        WHEN status = 'expired' THEN 'Domain not found'
        WHEN name LIKE '%dev%' OR name LIKE '%staging%' OR name LIKE '%testing%' THEN 'Service temporarily unavailable'
        WHEN name LIKE '%premium%' OR name LIKE '%future%' THEN 'Redirected to landing page'
        WHEN name LIKE '%blockchain%' THEN 'Access forbidden'
        ELSE 'OK'
    END;

-- Insert comprehensive DNS records for demo domains
DO $$
DECLARE
    domain_rec RECORD;
    a_ip TEXT;
    mail_server TEXT;
BEGIN
    -- Loop through all active domains to create DNS records
    FOR domain_rec IN SELECT id, name FROM domains WHERE status = 'active' LOOP
        -- Generate realistic IP addresses based on domain type
        a_ip := CASE 
            WHEN domain_rec.name LIKE '%tech%' OR domain_rec.name LIKE '%solutions%' THEN '192.168.1.10'
            WHEN domain_rec.name LIKE '%photography%' OR domain_rec.name LIKE '%boutique%' THEN '203.0.113.15'
            WHEN domain_rec.name LIKE '%dev%' OR domain_rec.name LIKE '%staging%' THEN '10.0.0.5'
            WHEN domain_rec.name LIKE '%shop%' OR domain_rec.name LIKE '%store%' THEN '198.51.100.20'
            WHEN domain_rec.name LIKE '%premium%' OR domain_rec.name LIKE '%future%' THEN '172.16.0.10'
            ELSE '203.0.113.25'
        END;
        
        mail_server := CASE 
            WHEN domain_rec.name LIKE '%tech%' OR domain_rec.name LIKE '%solutions%' THEN 'mail.' || domain_rec.name
            WHEN domain_rec.name LIKE '%photography%' OR domain_rec.name LIKE '%boutique%' THEN 'mx1.emailprovider.com'
            WHEN domain_rec.name LIKE '%shop%' OR domain_rec.name LIKE '%store%' THEN 'mail.shopify.com'
            ELSE 'mx.google.com'
        END;
        
        -- Insert A record for root domain
        INSERT INTO dns_records (id, domain_id, type, name, value, ttl, created_at, updated_at) VALUES
        (gen_random_uuid(), domain_rec.id, 'A', '@', a_ip, 3600, NOW(), NOW());
        
        -- Insert A record for www subdomain
        INSERT INTO dns_records (id, domain_id, type, name, value, ttl, created_at, updated_at) VALUES
        (gen_random_uuid(), domain_rec.id, 'A', 'www', a_ip, 3600, NOW(), NOW());
        
        -- Insert CNAME for common subdomains based on domain type
        IF domain_rec.name LIKE '%tech%' OR domain_rec.name LIKE '%solutions%' OR domain_rec.name LIKE '%saas%' THEN
            -- Tech domains get API and admin subdomains
            INSERT INTO dns_records (id, domain_id, type, name, value, ttl, created_at, updated_at) VALUES
            (gen_random_uuid(), domain_rec.id, 'CNAME', 'api', domain_rec.name, 3600, NOW(), NOW()),
            (gen_random_uuid(), domain_rec.id, 'CNAME', 'admin', domain_rec.name, 3600, NOW(), NOW()),
            (gen_random_uuid(), domain_rec.id, 'CNAME', 'app', domain_rec.name, 3600, NOW(), NOW());
        END IF;
        
        IF domain_rec.name LIKE '%shop%' OR domain_rec.name LIKE '%store%' OR domain_rec.name LIKE '%boutique%' THEN
            -- E-commerce domains get store-related subdomains
            INSERT INTO dns_records (id, domain_id, type, name, value, ttl, created_at, updated_at) VALUES
            (gen_random_uuid(), domain_rec.id, 'CNAME', 'shop', domain_rec.name, 3600, NOW(), NOW()),
            (gen_random_uuid(), domain_rec.id, 'CNAME', 'checkout', domain_rec.name, 3600, NOW(), NOW()),
            (gen_random_uuid(), domain_rec.id, 'CNAME', 'cdn', 'cdn.cloudflare.com', 300, NOW(), NOW());
        END IF;
        
        IF domain_rec.name LIKE '%dev%' OR domain_rec.name LIKE '%staging%' OR domain_rec.name LIKE '%testing%' THEN
            -- Development domains get dev-related subdomains
            INSERT INTO dns_records (id, domain_id, type, name, value, ttl, created_at, updated_at) VALUES
            (gen_random_uuid(), domain_rec.id, 'CNAME', 'staging', domain_rec.name, 300, NOW(), NOW()),
            (gen_random_uuid(), domain_rec.id, 'CNAME', 'dev', domain_rec.name, 300, NOW(), NOW()),
            (gen_random_uuid(), domain_rec.id, 'CNAME', 'test', domain_rec.name, 300, NOW(), NOW());
        END IF;
        
        -- Insert MX record for email
        INSERT INTO dns_records (id, domain_id, type, name, value, ttl, priority, created_at, updated_at) VALUES
        (gen_random_uuid(), domain_rec.id, 'MX', '@', mail_server, 3600, 10, NOW(), NOW());
        
        -- Insert additional MX record for backup (some domains)
        IF domain_rec.name LIKE '%tech%' OR domain_rec.name LIKE '%business%' THEN
            INSERT INTO dns_records (id, domain_id, type, name, value, ttl, priority, created_at, updated_at) VALUES
            (gen_random_uuid(), domain_rec.id, 'MX', '@', 'backup.' || mail_server, 3600, 20, NOW(), NOW());
        END IF;
        
        -- Insert TXT records for various purposes
        INSERT INTO dns_records (id, domain_id, type, name, value, ttl, created_at, updated_at) VALUES
        (gen_random_uuid(), domain_rec.id, 'TXT', '@', 'v=spf1 include:_spf.google.com ~all', 3600, NOW(), NOW()),
        (gen_random_uuid(), domain_rec.id, 'TXT', '_dmarc', 'v=DMARC1; p=quarantine; rua=mailto:dmarc@' || domain_rec.name, 3600, NOW(), NOW());
        
        -- Insert AAAA record for IPv6 (some domains)
        IF domain_rec.name LIKE '%tech%' OR domain_rec.name LIKE '%future%' OR domain_rec.name LIKE '%innovation%' THEN
            INSERT INTO dns_records (id, domain_id, type, name, value, ttl, created_at, updated_at) VALUES
            (gen_random_uuid(), domain_rec.id, 'AAAA', '@', '2001:db8::1', 3600, NOW(), NOW()),
            (gen_random_uuid(), domain_rec.id, 'AAAA', 'www', '2001:db8::1', 3600, NOW(), NOW());
        END IF;
        
        -- Insert SRV records for some business domains
        IF domain_rec.name LIKE '%tech%' OR domain_rec.name LIKE '%solutions%' THEN
            INSERT INTO dns_records (id, domain_id, type, name, value, ttl, priority, weight, port, created_at, updated_at) VALUES
            (gen_random_uuid(), domain_rec.id, 'SRV', '_sip._tcp', 'sip.' || domain_rec.name, 3600, 10, 5, 5060, NOW(), NOW()),
            (gen_random_uuid(), domain_rec.id, 'SRV', '_xmpp-server._tcp', 'xmpp.' || domain_rec.name, 3600, 5, 0, 5269, NOW(), NOW());
        END IF;
        
        -- Insert CAA records for certificate authority authorization
        IF domain_rec.name LIKE '%tech%' OR domain_rec.name LIKE '%shop%' OR domain_rec.name LIKE '%boutique%' THEN
            INSERT INTO dns_records (id, domain_id, type, name, value, ttl, created_at, updated_at) VALUES
            (gen_random_uuid(), domain_rec.id, 'CAA', '@', '0 issue "letsencrypt.org"', 3600, NOW(), NOW()),
            (gen_random_uuid(), domain_rec.id, 'CAA', '@', '0 iodef "mailto:security@' || domain_rec.name || '"', 3600, NOW(), NOW());
        END IF;
        
    END LOOP;
END $$;

-- Update statistics
ANALYZE dns_records;

-- Show summary of DNS records created
SELECT 
    'DNS records created' as item,
    COUNT(*) as count
FROM dns_records;

-- Show DNS records by type
SELECT 
    type,
    COUNT(*) as count
FROM dns_records
GROUP BY type
ORDER BY count DESC;

-- Show domains with their HTTP status
SELECT 
    name,
    http_status,
    status_message,
    last_status_check
FROM domains
ORDER BY http_status, name;

-- Show sample DNS records
SELECT 
    d.name as domain,
    dr.type,
    dr.name as record_name,
    dr.value,
    dr.ttl,
    dr.priority
FROM domains d
JOIN dns_records dr ON d.id = dr.domain_id
WHERE d.name IN ('techsolutions.com', 'fashionboutique.com', 'jsmith.dev')
ORDER BY d.name, dr.type, dr.name;
