#!/bin/bash

# DomainVault Production Server Startup Script
echo "🚀 Starting DomainVault Production Server"
echo "=========================================="

# Check if PostgreSQL is running
if ! brew services list | grep postgresql@15 | grep started > /dev/null; then
    echo "📊 Starting PostgreSQL..."
    brew services start postgresql@15
    sleep 2
fi

# Check database connection
echo "🔗 Testing database connection..."
if ! /opt/homebrew/opt/postgresql@15/bin/psql domainvault -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ Database connection failed. Please check PostgreSQL installation."
    exit 1
fi

echo "✅ Database connected successfully"

# Set environment variables
export DATABASE_URL="postgres://domainvault_user:domainvault_pass@localhost/domainvault?sslmode=disable"
export PORT=8080
export GIN_MODE=release
export LOG_LEVEL=info

echo "📋 Configuration:"
echo "  Database: PostgreSQL (localhost:5432/domainvault)"
echo "  Port: $PORT"
echo "  Environment: production"
echo ""

echo "🔧 Starting server..."
echo "📱 Admin interface: http://localhost:$PORT/admin"
echo "🔑 Admin credentials: admin / admin123"
echo "🌐 API endpoints: http://localhost:$PORT/api/v1/"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Start the server
go run cmd/server/main.go
