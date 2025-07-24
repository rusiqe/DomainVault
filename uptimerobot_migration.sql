-- UptimeRobot Integration Migration
-- This script adds UptimeRobot monitoring fields to the domains table
-- and creates a configuration table for UptimeRobot settings

-- Add UptimeRobot fields to domains table
ALTER TABLE domains ADD COLUMN IF NOT EXISTS uptime_robot_monitor_id INTEGER;
ALTER TABLE domains ADD COLUMN IF NOT EXISTS uptime_ratio DECIMAL(5,2);
ALTER TABLE domains ADD COLUMN IF NOT EXISTS response_time INTEGER;
ALTER TABLE domains ADD COLUMN IF NOT EXISTS monitor_status VARCHAR(20);
ALTER TABLE domains ADD COLUMN IF NOT EXISTS last_downtime TIMESTAMPTZ;

-- Create indexes for the new fields
CREATE INDEX IF NOT EXISTS idx_domains_uptime_robot_monitor_id ON domains(uptime_robot_monitor_id);
CREATE INDEX IF NOT EXISTS idx_domains_monitor_status ON domains(monitor_status);
CREATE INDEX IF NOT EXISTS idx_domains_uptime_ratio ON domains(uptime_ratio);

-- Create UptimeRobot configuration table
CREATE TABLE IF NOT EXISTS uptimerobot_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key TEXT NOT NULL,
    enabled BOOLEAN DEFAULT false,
    interval INTEGER DEFAULT 300, -- Default 5 minutes
    alert_contacts TEXT[], -- Array of alert contact IDs
    auto_create_monitors BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create UptimeRobot monitor log table for tracking monitor history
CREATE TABLE IF NOT EXISTS uptimerobot_monitor_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain_id UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    monitor_id INTEGER NOT NULL,
    event_type VARCHAR(20) NOT NULL, -- created, updated, deleted, paused, resumed
    old_status VARCHAR(20),
    new_status VARCHAR(20),
    uptime_ratio DECIMAL(5,2),
    response_time INTEGER,
    message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for the monitor logs table
CREATE INDEX IF NOT EXISTS idx_uptimerobot_monitor_logs_domain_id ON uptimerobot_monitor_logs(domain_id);
CREATE INDEX IF NOT EXISTS idx_uptimerobot_monitor_logs_monitor_id ON uptimerobot_monitor_logs(monitor_id);
CREATE INDEX IF NOT EXISTS idx_uptimerobot_monitor_logs_event_type ON uptimerobot_monitor_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_uptimerobot_monitor_logs_created_at ON uptimerobot_monitor_logs(created_at);

-- Add comments to document the new fields
COMMENT ON COLUMN domains.uptime_robot_monitor_id IS 'UptimeRobot monitor ID for this domain';
COMMENT ON COLUMN domains.uptime_ratio IS 'Current uptime ratio percentage (0-100)';
COMMENT ON COLUMN domains.response_time IS 'Average response time in milliseconds';
COMMENT ON COLUMN domains.monitor_status IS 'Current monitor status: up, down, paused, seems_down';
COMMENT ON COLUMN domains.last_downtime IS 'Timestamp of last recorded downtime';

COMMENT ON TABLE uptimerobot_config IS 'UptimeRobot API configuration settings';
COMMENT ON TABLE uptimerobot_monitor_logs IS 'Log of UptimeRobot monitor events and status changes';

-- Create a function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for uptimerobot_config table
CREATE TRIGGER update_uptimerobot_config_updated_at
    BEFORE UPDATE ON uptimerobot_config
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default configuration (disabled by default)
INSERT INTO uptimerobot_config (api_key, enabled, interval, auto_create_monitors)
VALUES ('', false, 300, true)
ON CONFLICT DO NOTHING;

-- Create a view for domain monitoring summary
CREATE OR REPLACE VIEW domain_monitoring_summary AS
SELECT 
    d.id,
    d.name,
    d.provider,
    d.status as domain_status,
    d.expires_at,
    d.http_status,
    d.last_status_check,
    d.status_message,
    d.uptime_robot_monitor_id,
    d.uptime_ratio,
    d.response_time,
    d.monitor_status,
    d.last_downtime,
    CASE 
        WHEN d.uptime_robot_monitor_id IS NOT NULL THEN 'monitored'
        ELSE 'not_monitored'
    END as monitoring_status,
    CASE
        WHEN d.monitor_status = 'up' THEN 'healthy'
        WHEN d.monitor_status IN ('down', 'seems_down') THEN 'unhealthy'
        WHEN d.monitor_status = 'paused' THEN 'paused'
        ELSE 'unknown'
    END as health_status
FROM domains d
ORDER BY d.name;

COMMENT ON VIEW domain_monitoring_summary IS 'Summary view of domain monitoring status including UptimeRobot data';

-- Create a function to get monitoring statistics
CREATE OR REPLACE FUNCTION get_monitoring_stats()
RETURNS TABLE (
    total_domains INTEGER,
    monitored_domains INTEGER,
    unmonitored_domains INTEGER,
    up_domains INTEGER,
    down_domains INTEGER,
    paused_domains INTEGER,
    avg_uptime_ratio DECIMAL(5,2),
    avg_response_time INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*)::INTEGER as total_domains,
        COUNT(uptime_robot_monitor_id)::INTEGER as monitored_domains,
        (COUNT(*) - COUNT(uptime_robot_monitor_id))::INTEGER as unmonitored_domains,
        COUNT(CASE WHEN monitor_status = 'up' THEN 1 END)::INTEGER as up_domains,
        COUNT(CASE WHEN monitor_status IN ('down', 'seems_down') THEN 1 END)::INTEGER as down_domains,
        COUNT(CASE WHEN monitor_status = 'paused' THEN 1 END)::INTEGER as paused_domains,
        ROUND(AVG(uptime_ratio), 2) as avg_uptime_ratio,
        ROUND(AVG(response_time))::INTEGER as avg_response_time
    FROM domains;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_monitoring_stats() IS 'Returns aggregated monitoring statistics for all domains';

-- Update demo data with some UptimeRobot monitoring data (optional)
-- This adds sample monitoring data to some existing domains
UPDATE domains 
SET 
    uptime_robot_monitor_id = 123456789 + ROW_NUMBER() OVER (ORDER BY name),
    uptime_ratio = 99.5 + (RANDOM() * 0.5 - 0.25), -- Random between 99.25 and 99.75
    response_time = 150 + (RANDOM() * 200)::INTEGER, -- Random between 150-350ms
    monitor_status = CASE 
        WHEN RANDOM() < 0.85 THEN 'up'
        WHEN RANDOM() < 0.95 THEN 'seems_down'
        ELSE 'down'
    END
WHERE name IN (
    'techsolutions.com',
    'innovatetech.io',
    'bestdeals.shop',
    'creativestudio.net',
    'fashionboutique.com',
    'greengardens.org',
    'smartfinance.co',
    'healthplus.info',
    'traveladventures.com',
    'foodiehaven.net'
);

-- Set some domains to have recent downtime
UPDATE domains 
SET last_downtime = NOW() - INTERVAL '2 hours'
WHERE monitor_status IN ('down', 'seems_down');

-- Print migration completion message
DO $$
BEGIN
    RAISE NOTICE 'UptimeRobot integration migration completed successfully!';
    RAISE NOTICE 'Added UptimeRobot monitoring fields to domains table';
    RAISE NOTICE 'Created uptimerobot_config table for API configuration';
    RAISE NOTICE 'Created uptimerobot_monitor_logs table for event tracking';
    RAISE NOTICE 'Created monitoring views and functions';
    RAISE NOTICE 'Updated sample domains with mock monitoring data';
    RAISE NOTICE '';
    RAISE NOTICE 'Next steps:';
    RAISE NOTICE '1. Configure your UptimeRobot API key in the uptimerobot_config table';
    RAISE NOTICE '2. Use the new API endpoints to manage UptimeRobot monitoring';
    RAISE NOTICE '3. View monitoring status using the domain_monitoring_summary view';
END $$;
