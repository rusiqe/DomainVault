-- Demo data for DomainVault
-- Insert sample domains with realistic data

-- Get category IDs for reference
DO $$
DECLARE
    personal_id UUID;
    business_id UUID;
    dev_id UUID;
    client_id UUID;
    investment_id UUID;
    portfolio_id UUID;
    ecommerce_id UUID;
    saas_id UUID;
    marketing_id UUID;
BEGIN
    -- Get category IDs
    SELECT id INTO personal_id FROM categories WHERE name = 'Personal';
    SELECT id INTO business_id FROM categories WHERE name = 'Business';
    SELECT id INTO dev_id FROM categories WHERE name = 'Development';
    SELECT id INTO client_id FROM categories WHERE name = 'Client Work';
    SELECT id INTO investment_id FROM categories WHERE name = 'Investment';
    
    -- Get project IDs
    SELECT id INTO portfolio_id FROM projects WHERE name = 'Portfolio Sites';
    SELECT id INTO ecommerce_id FROM projects WHERE name = 'E-commerce';
    SELECT id INTO saas_id FROM projects WHERE name = 'SaaS Products';
    SELECT id INTO marketing_id FROM projects WHERE name = 'Marketing Campaigns';

    -- Insert demo domains
    INSERT INTO domains (id, name, provider, expires_at, category_id, project_id, auto_renew, renewal_price, status, tags) VALUES
    -- Personal domains
    (gen_random_uuid(), 'johnsmith.com', 'godaddy', NOW() + INTERVAL '45 days', personal_id, portfolio_id, true, 15.99, 'active', '["personal", "portfolio"]'::jsonb),
    (gen_random_uuid(), 'john-smith-photography.com', 'namecheap', NOW() + INTERVAL '120 days', personal_id, portfolio_id, true, 12.88, 'active', '["photography", "personal"]'::jsonb),
    (gen_random_uuid(), 'jsmith.dev', 'godaddy', NOW() + INTERVAL '200 days', personal_id, portfolio_id, true, 25.99, 'active', '["development", "portfolio"]'::jsonb),
    
    -- Business domains
    (gen_random_uuid(), 'techsolutions.com', 'godaddy', NOW() + INTERVAL '15 days', business_id, saas_id, true, 15.99, 'active', '["business", "tech"]'::jsonb),
    (gen_random_uuid(), 'innovatetech.io', 'namecheap', NOW() + INTERVAL '90 days', business_id, saas_id, true, 45.99, 'active', '["startup", "tech"]'::jsonb),
    (gen_random_uuid(), 'cloudsolutions.net', 'godaddy', NOW() + INTERVAL '180 days', business_id, saas_id, true, 18.99, 'active', '["cloud", "enterprise"]'::jsonb),
    
    -- E-commerce domains
    (gen_random_uuid(), 'bestdeals.shop', 'namecheap', NOW() + INTERVAL '60 days', business_id, ecommerce_id, true, 35.99, 'active', '["ecommerce", "retail"]'::jsonb),
    (gen_random_uuid(), 'fashionboutique.com', 'godaddy', NOW() + INTERVAL '300 days', business_id, ecommerce_id, true, 15.99, 'active', '["fashion", "retail"]'::jsonb),
    (gen_random_uuid(), 'gadgetstore.online', 'namecheap', NOW() + INTERVAL '150 days', business_id, ecommerce_id, true, 28.99, 'active', '["electronics", "gadgets"]'::jsonb),
    
    -- Development/Client domains
    (gen_random_uuid(), 'client-project-alpha.com', 'godaddy', NOW() + INTERVAL '30 days', client_id, NULL, false, 15.99, 'active', '["client", "project"]'::jsonb),
    (gen_random_uuid(), 'beta-testing.dev', 'namecheap', NOW() + INTERVAL '75 days', dev_id, NULL, true, 32.99, 'active', '["testing", "development"]'::jsonb),
    (gen_random_uuid(), 'staging-environment.net', 'godaddy', NOW() + INTERVAL '100 days', dev_id, NULL, true, 18.99, 'active', '["staging", "development"]'::jsonb),
    
    -- Investment domains
    (gen_random_uuid(), 'premiumdomain.com', 'namecheap', NOW() + INTERVAL '250 days', investment_id, NULL, true, 299.99, 'active', '["premium", "investment"]'::jsonb),
    (gen_random_uuid(), 'future-tech.ai', 'godaddy', NOW() + INTERVAL '180 days', investment_id, NULL, true, 199.99, 'active', '["ai", "future"]'::jsonb),
    (gen_random_uuid(), 'blockchain-solutions.crypto', 'namecheap', NOW() + INTERVAL '365 days', investment_id, NULL, true, 89.99, 'active', '["blockchain", "crypto"]'::jsonb),
    
    -- Marketing campaigns
    (gen_random_uuid(), 'summer-sale-2024.com', 'godaddy', NOW() + INTERVAL '20 days', business_id, marketing_id, false, 15.99, 'active', '["campaign", "seasonal"]'::jsonb),
    (gen_random_uuid(), 'product-launch.net', 'namecheap', NOW() + INTERVAL '85 days', business_id, marketing_id, true, 22.99, 'active', '["launch", "marketing"]'::jsonb),
    
    -- Some expiring soon (for testing alerts)
    (gen_random_uuid(), 'urgent-renewal.com', 'godaddy', NOW() + INTERVAL '5 days', business_id, NULL, false, 15.99, 'active', '["urgent", "expiring"]'::jsonb),
    (gen_random_uuid(), 'almost-expired.org', 'namecheap', NOW() + INTERVAL '12 days', personal_id, NULL, true, 18.99, 'active', '["expiring"]'::jsonb),
    
    -- Some expired (for testing)
    (gen_random_uuid(), 'old-project.com', 'godaddy', NOW() - INTERVAL '10 days', dev_id, NULL, false, 15.99, 'expired', '["old", "expired"]'::jsonb),
    (gen_random_uuid(), 'abandoned-site.net', 'namecheap', NOW() - INTERVAL '30 days', personal_id, NULL, false, 12.88, 'expired', '["abandoned"]'::jsonb);

END $$;

-- Update statistics
ANALYZE domains;
ANALYZE categories;
ANALYZE projects;

-- Show summary of inserted data
SELECT 
    'Domains inserted' as item,
    COUNT(*) as count
FROM domains
UNION ALL
SELECT 
    'Categories available' as item,
    COUNT(*) as count
FROM categories
UNION ALL
SELECT 
    'Projects available' as item,
    COUNT(*) as count
FROM projects;

-- Show domains by status
SELECT 
    status,
    COUNT(*) as count
FROM domains
GROUP BY status
ORDER BY status;

-- Show domains expiring in next 30 days
SELECT 
    name,
    provider,
    expires_at,
    EXTRACT(days FROM expires_at - NOW()) as days_until_expiry
FROM domains
WHERE expires_at BETWEEN NOW() AND NOW() + INTERVAL '30 days'
ORDER BY expires_at;
