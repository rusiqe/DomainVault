#!/bin/bash

# DomainVault Environment Setup Script
# This script helps you set up your environment variables securely

set -e

echo "🏰 DomainVault Environment Setup"
echo "================================"
echo

# Function to read input securely
read_input() {
    local prompt="$1"
    local var_name="$2"
    local is_secret="$3"
    local default_value="$4"
    
    if [ "$is_secret" = "true" ]; then
        echo -n "$prompt: " >&2
        read -s input
        echo >&2
    else
        if [ -n "$default_value" ]; then
            echo -n "$prompt [$default_value]: " >&2
        else
            echo -n "$prompt: " >&2
        fi
        read input
        if [ -z "$input" ] && [ -n "$default_value" ]; then
            input="$default_value"
        fi
    fi
    
    eval "$var_name='$input'"
}

# Check if .env already exists
if [ -f ".env" ]; then
    echo "⚠️  Found existing .env file!"
    echo -n "Do you want to backup the existing .env file? (y/N): "
    read backup_choice
    if [[ $backup_choice =~ ^[Yy]$ ]]; then
        cp .env ".env.backup.$(date +%Y%m%d_%H%M%S)"
        echo "✅ Backed up existing .env file"
    fi
    echo
fi

echo "Let's set up your environment variables..."
echo

# Database Configuration
echo "📊 Database Configuration"
echo "-------------------------"
read_input "Database URL" DB_URL false "postgres://user:password@localhost/domainvault?sslmode=disable"
echo

# Server Configuration
echo "🖥️  Server Configuration"
echo "------------------------"
read_input "Server Port" SERVER_PORT false "8080"
read_input "Log Level (debug, info, warn, error)" LOG_LEVEL false "info"
echo

# UptimeRobot Configuration
echo "🤖 UptimeRobot Configuration"
echo "----------------------------"
echo "To get your UptimeRobot API key:"
echo "1. Sign up at https://uptimerobot.com/"
echo "2. Go to Dashboard → My Settings → API Settings"
echo "3. Create a 'Main API Key' (not 'Monitor-specific API Keys')"
echo

read_input "UptimeRobot API Key" UPTIMEROBOT_KEY true
read_input "Enable UptimeRobot monitoring (true/false)" UPTIMEROBOT_ENABLED false "true"
read_input "Monitor check interval in seconds (300, 600, 900, 1800, 3600)" UPTIMEROBOT_INTERVAL false "300"
read_input "Alert Contact IDs (comma-separated, optional)" UPTIMEROBOT_CONTACTS false ""
read_input "Auto-create monitors for new domains (true/false)" UPTIMEROBOT_AUTO_CREATE false "true"
echo

# Domain Registrar APIs (Optional)
echo "🌐 Domain Registrar APIs (Optional)"
echo "-----------------------------------"
echo "These are optional. Press Enter to skip if you don't have them."
echo

echo "GoDaddy API (https://developer.godaddy.com/keys):"
read_input "GoDaddy API Key (optional)" GODADDY_KEY false ""
if [ -n "$GODADDY_KEY" ]; then
    read_input "GoDaddy API Secret" GODADDY_SECRET true
fi
echo

echo "Namecheap API (https://www.namecheap.com/support/api/intro/):"
read_input "Namecheap API Key (optional)" NAMECHEAP_KEY false ""
if [ -n "$NAMECHEAP_KEY" ]; then
    read_input "Namecheap Username" NAMECHEAP_USER false ""
fi
echo

# Create .env file
echo "📝 Creating .env file..."

cat > .env << EOF
# DomainVault Environment Configuration
# Generated on $(date)

# ================================
# SERVER CONFIGURATION
# ================================
PORT=$SERVER_PORT
LOG_LEVEL=$LOG_LEVEL
SYNC_INTERVAL=1h

# ================================
# DATABASE CONFIGURATION
# ================================
DATABASE_URL=$DB_URL

# ================================
# UPTIMEROBOT MONITORING
# ================================
UPTIMEROBOT_API_KEY=$UPTIMEROBOT_KEY
UPTIMEROBOT_ENABLED=$UPTIMEROBOT_ENABLED
UPTIMEROBOT_INTERVAL=$UPTIMEROBOT_INTERVAL
UPTIMEROBOT_ALERT_CONTACTS=$UPTIMEROBOT_CONTACTS
UPTIMEROBOT_AUTO_CREATE=$UPTIMEROBOT_AUTO_CREATE

EOF

# Add registrar configs if provided
if [ -n "$GODADDY_KEY" ]; then
    cat >> .env << EOF

# ================================
# GODADDY API CONFIGURATION
# ================================
GODADDY_API_KEY=$GODADDY_KEY
GODADDY_API_SECRET=$GODADDY_SECRET
EOF
fi

if [ -n "$NAMECHEAP_KEY" ]; then
    cat >> .env << EOF

# ================================
# NAMECHEAP API CONFIGURATION
# ================================
NAMECHEAP_API_KEY=$NAMECHEAP_KEY
NAMECHEAP_USERNAME=$NAMECHEAP_USER
EOF
fi

# Set proper permissions
chmod 600 .env

echo "✅ Environment file created successfully!"
echo
echo "🔒 Security Notes:"
echo "- The .env file has been created with restricted permissions (600)"
echo "- Never commit the .env file to version control"
echo "- Keep your API keys secure and rotate them regularly"
echo
echo "🚀 Next Steps:"
echo "1. Run 'go run main.go' to start DomainVault"
echo "2. Visit http://localhost:$SERVER_PORT/admin to access the admin panel"
echo "3. Use the monitoring endpoints to manage UptimeRobot integration"
echo
echo "📖 API Endpoints:"
echo "- GET  /api/v1/monitoring/stats  - Get monitoring statistics"
echo "- POST /api/v1/monitoring/sync   - Sync all domain monitoring"
echo "- POST /api/v1/monitoring/create - Create monitors for specific domains"
echo
echo "Happy monitoring! 🎉"
