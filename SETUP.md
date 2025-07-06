# DomainVault Setup Guide

## ✅ Current Status
The DomainVault application is now fully set up with:
- PostgreSQL database with schema and demo data
- Working admin authentication
- 21 demo domains with realistic data
- Categories and projects for organization
- **157 DNS records** across all domains with realistic configurations
- **HTTP status monitoring** with live status checking capabilities

## 🔑 Admin Credentials
- **Username**: `admin`
- **Password**: `admin123`

## 🚀 Quick Start

### Start the Server
```bash
./start_prod.sh
```

### Access the Application
- **Admin Interface**: http://localhost:8080/admin
- **Main Dashboard**: http://localhost:8080/
- **API Health Check**: http://localhost:8080/api/v1/health

## 📊 Demo Data
The database includes 21 sample domains across different categories:

### Domain Status Distribution
- **Active**: 19 domains
- **Expired**: 2 domains

### Provider Distribution  
- **GoDaddy**: 11 domains
- **Namecheap**: 10 domains

### Expiration Timeline
- **Next 30 days**: 5 domains expiring
- **Next 90 days**: 10 domains expiring
- **Next 365 days**: 19 domains expiring

### DNS Records
- **Total Records**: 157 DNS records
- **A Records**: 38 (IPv4 addresses)
- **TXT Records**: 38 (SPF, DMARC, etc.)
- **CNAME Records**: 33 (Subdomains)
- **MX Records**: 22 (Email routing)
- **CAA Records**: 10 (Certificate authority)
- **SRV Records**: 10 (Service discovery)
- **AAAA Records**: 6 (IPv6 addresses)

### HTTP Status Monitoring
- **200 (OK)**: Most active domains
- **301 (Redirect)**: Premium/parked domains
- **404 (Not Found)**: Expired domains
- **503 (Unavailable)**: Development domains
- **Last Checked**: Real-time status monitoring

### Categories
- Personal (3 domains)
- Business (8 domains) 
- Development (3 domains)
- Client Work (1 domain)
- Investment (3 domains)

### Projects
- Portfolio Sites
- E-commerce
- SaaS Products
- Marketing Campaigns

## 🗄️ Database Details
- **Database Name**: `domainvault`
- **User**: `domainvault_user`
- **Password**: `domainvault_pass`
- **Host**: `localhost:5432`

## 📋 API Endpoints

### Authentication
```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'
```

### Domains
```bash
# Get all domains
curl http://localhost:8080/api/v1/domains

# Get domain summary
curl http://localhost:8080/api/v1/domains/summary

# Get expiring domains
curl http://localhost:8080/api/v1/domains/expiring?days=30
```

### DNS Management
```bash
# Get DNS records for a domain
curl http://localhost:8080/api/v1/admin/domains/{domain_id}/dns \
  -H "Authorization: Bearer {token}"
```

### Status Monitoring
```bash
# Check status of a single domain
curl -X POST http://localhost:8080/api/v1/admin/domains/{domain_id}/check-status \
  -H "Authorization: Bearer {token}"

# Bulk check multiple domains
curl -X POST http://localhost:8080/api/v1/admin/domains/bulk-check-status \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"domain_ids": ["id1", "id2"], "check_https": true}'

# Get status summary
curl http://localhost:8080/api/v1/admin/status/summary \
  -H "Authorization: Bearer {token}"
```

### Health Check
```bash
curl http://localhost:8080/api/v1/health
```

## 🛠️ Development Commands

### Database Management
```bash
# Connect to database
/opt/homebrew/opt/postgresql@15/bin/psql domainvault

# Reset demo data
/opt/homebrew/opt/postgresql@15/bin/psql domainvault -f setup_demo_data.sql

# View all domains
/opt/homebrew/opt/postgresql@15/bin/psql domainvault -c "SELECT name, provider, expires_at FROM domains ORDER BY expires_at;"
```

### Server Management
```bash
# Start production server
./start_prod.sh

# Start development server (in-memory)
go run cmd/dev/main.go

# Check what's running on port 8080
lsof -i :8080
```

## 🔧 Troubleshooting

### Login Issues
If `admin/admin123` doesn't work:
1. Check the server is running: `curl http://localhost:8080/api/v1/health`
2. Verify admin user exists: `/opt/homebrew/opt/postgresql@15/bin/psql domainvault -c "SELECT username FROM users;"`
3. Check server logs for authentication errors

### Database Issues
If database connection fails:
1. Start PostgreSQL: `brew services start postgresql@15`
2. Test connection: `/opt/homebrew/opt/postgresql@15/bin/psql domainvault -c "SELECT 1;"`
3. Check if database exists: `/opt/homebrew/opt/postgresql@15/bin/psql -l | grep domainvault`

### Port Conflicts
If port 8080 is in use:
1. Find the process: `lsof -i :8080`
2. Kill the process: `kill <PID>`
3. Or use a different port: `PORT=8081 ./start_prod.sh`

## 📝 Notes
- The demo data includes domains with various expiration dates for testing alerts
- Some domains are intentionally expired to test the expired domain functionality
- All domains use realistic pricing and categorization
- The admin interface is fully functional with the demo data
