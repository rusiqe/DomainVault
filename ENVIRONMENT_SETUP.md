# 🔐 DomainVault Environment Setup Guide

This guide explains how to securely configure DomainVault with environment variables for API keys and sensitive configuration.

## 🚀 Quick Setup

### Option 1: Interactive Setup Script (Recommended)
```bash
./setup_env.sh
```

This script will guide you through setting up all environment variables securely.

### Option 2: Manual Setup
```bash
# Copy the example file and edit it
cp .env.example .env
nano .env  # or use your preferred editor
```

## 📋 Required Environment Variables

### Server Configuration
```bash
PORT=8080                    # Server port
LOG_LEVEL=info              # Log level (debug, info, warn, error)
SYNC_INTERVAL=1h            # How often to sync domains
DATABASE_URL=postgres://... # PostgreSQL connection string
```

### UptimeRobot Configuration
```bash
UPTIMEROBOT_API_KEY=ur123456789...    # Your UptimeRobot API key
UPTIMEROBOT_ENABLED=true              # Enable/disable monitoring
UPTIMEROBOT_INTERVAL=300              # Check interval (seconds)
UPTIMEROBOT_ALERT_CONTACTS=12345,678  # Comma-separated contact IDs
UPTIMEROBOT_AUTO_CREATE=true          # Auto-create monitors
```

### Domain Registrar APIs (Optional)
```bash
# GoDaddy
GODADDY_API_KEY=your_key
GODADDY_API_SECRET=your_secret

# Namecheap
NAMECHEAP_API_KEY=your_key
NAMECHEAP_USERNAME=your_username
```

## 🔑 Getting API Keys

### UptimeRobot API Key
1. Sign up at [UptimeRobot](https://uptimerobot.com/)
2. Go to **Dashboard → My Settings → API Settings**
3. Create a **"Main API Key"** (not "Monitor-specific API Keys")
4. Copy the API key to `UPTIMEROBOT_API_KEY`

### UptimeRobot Alert Contacts (Optional)
1. Go to **Dashboard → My Settings → Alert Contacts**
2. Set up your email, SMS, Slack, etc.
3. Note the ID numbers of each contact
4. Add them to `UPTIMEROBOT_ALERT_CONTACTS` as comma-separated values

Example: `UPTIMEROBOT_ALERT_CONTACTS=12345,67890,11111`

### GoDaddy API Keys (Optional)
1. Visit [GoDaddy Developer Portal](https://developer.godaddy.com/keys)
2. Create production or development keys
3. Copy API Key and Secret

### Namecheap API Keys (Optional)
1. Visit [Namecheap API Documentation](https://www.namecheap.com/support/api/intro/)
2. Enable API access in your account
3. Get your API key and username

## 🛡️ Security Best Practices

### File Permissions
```bash
# Set restrictive permissions on .env file
chmod 600 .env
```

### Version Control
The `.gitignore` file already excludes:
- `.env`
- `.env.*`
- All environment files

**Never commit API keys to version control!**

### Environment-Specific Files
```bash
.env                    # Local development (ignored by git)
.env.production        # Production (ignored by git)
.env.staging          # Staging (ignored by git)
.env.example          # Template (committed to git)
```

### Key Rotation
- Rotate API keys regularly
- Use different keys for development/production
- Monitor API key usage in provider dashboards

## 🔧 Testing Configuration

### Test Environment Loading
```bash
# Check if environment variables are loaded
go run -c "
package main
import (
    \"fmt\"
    \"github.com/rusiqe/domainvault/internal/config\"
)
func main() {
    cfg, err := config.Load()
    if err != nil {
        fmt.Printf(\"Error: %v\\n\", err)
        return
    }
    fmt.Printf(\"Port: %d\\n\", cfg.Port)
    if cfg.UptimeRobot != nil {
        fmt.Printf(\"UptimeRobot: %s\\n\", cfg.UptimeRobot.Enabled)
    }
}
"
```

### Test UptimeRobot Connection
```bash
curl -X GET http://localhost:8080/api/v1/monitoring/stats
```

## 🌍 Production Deployment

### Docker Environment
```dockerfile
# Dockerfile
ENV PORT=8080
ENV DATABASE_URL=${DATABASE_URL}
ENV UPTIMEROBOT_API_KEY=${UPTIMEROBOT_API_KEY}
ENV UPTIMEROBOT_ENABLED=true
```

### Docker Compose
```yaml
# docker-compose.yml
services:
  domainvault:
    environment:
      - PORT=8080
      - DATABASE_URL=${DATABASE_URL}
      - UPTIMEROBOT_API_KEY=${UPTIMEROBOT_API_KEY}
      - UPTIMEROBOT_ENABLED=true
    env_file:
      - .env.production
```

### Kubernetes Secrets
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: domainvault-secrets
type: Opaque
stringData:
  uptimerobot-api-key: "ur123456789..."
  database-url: "postgres://..."
```

### Cloud Provider Secrets
- **AWS**: Use AWS Secrets Manager or Parameter Store
- **Azure**: Use Azure Key Vault
- **GCP**: Use Google Secret Manager
- **Heroku**: Use Config Vars

## 🚨 Troubleshooting

### Common Issues

#### "UptimeRobot is not configured"
- Check `UPTIMEROBOT_API_KEY` is set
- Verify `UPTIMEROBOT_ENABLED=true`
- Test API key validity

#### "Database connection failed"
- Verify `DATABASE_URL` format
- Check database is running
- Verify credentials and permissions

#### "Invalid interval"
UptimeRobot intervals must be one of:
- `60`, `120`, `300`, `600`, `900`, `1800`, `3600` seconds

### Debug Configuration
```bash
# Print loaded configuration (without secrets)
go run main.go --debug-config
```

### Environment Variable Precedence
1. System environment variables (highest priority)
2. `.env` file
3. Default values (lowest priority)

## 📖 API Usage Examples

### Get Monitoring Statistics
```bash
curl -X GET http://localhost:8080/api/v1/monitoring/stats
```

### Sync All Domain Monitoring
```bash
curl -X POST http://localhost:8080/api/v1/monitoring/sync
```

### Create Monitors for Specific Domains
```bash
curl -X POST http://localhost:8080/api/v1/monitoring/create \
  -H "Content-Type: application/json" \
  -d '{
    "domain_ids": ["domain-uuid-1", "domain-uuid-2"],
    "monitor_type": "http",
    "interval": 300
  }'
```

## 🆘 Support

If you encounter issues:
1. Check this documentation
2. Verify your API keys are valid
3. Check the application logs
4. Test individual components

Remember: **Never share your API keys publicly!** 🔒
