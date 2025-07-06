#!/bin/bash

# DomainVault Development Server
echo "🚀 Starting DomainVault in development mode..."

# Set environment variables for development
export PORT=8080
export LOG_LEVEL=info
export SYNC_INTERVAL=5m

# Use SQLite for development (we'll need to modify the code slightly)
# For now, let's just run with mock database
export DATABASE_URL="mock://localhost/domainvault"

# Configure mock provider (no real API keys needed)
export MOCK_PROVIDER=true

echo "📋 Configuration:"
echo "  Port: $PORT"
echo "  Log Level: $LOG_LEVEL" 
echo "  Sync Interval: $SYNC_INTERVAL"
echo "  Database: Mock (for development)"
echo "  Provider: Mock (generates sample data)"
echo ""

echo "🔧 Starting server..."
go run cmd/server/main.go
