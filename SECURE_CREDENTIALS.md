# Secure Credential Management

DomainVault now supports secure credential management that stores API keys in environment variables instead of the database, significantly improving security.

## Overview

The new secure credential management system:

1. **Stores API keys in environment variables** (`.env` file) instead of the database
2. **Only stores references** to environment variable sets in the database
3. **Supports multiple environments** (production, staging) with different credential sets
4. **Enables easy credential rotation** without database changes
5. **Prevents credential exposure** in database backups or logs

## Security Benefits

### ✅ Improved Security
- API keys are never stored in the database
- Database backups don't contain sensitive credentials
- Credentials are managed at the environment level
- Easy to implement proper secrets management

### ✅ Environment Separation
- Different credentials for production/staging/development
- No risk of accidentally using production keys in development
- Clean separation of concerns

### ✅ Easy Credential Rotation
- Update credentials by changing environment variables only
- No database migrations needed for credential changes
- Zero-downtime credential updates

## Configuration

### Environment Variables

Add your provider credentials to the `.env` file:

```bash
# GoDaddy Production Account
GODADDY_API_KEY=your_production_api_key_here
GODADDY_API_SECRET=your_production_api_secret_here

# GoDaddy Staging Account (optional)
GODADDY_STAGING_API_KEY=your_staging_api_key_here  
GODADDY_STAGING_API_SECRET=your_staging_api_secret_here

# Namecheap Production Account
NAMECHEAP_API_KEY=your_namecheap_api_key_here
NAMECHEAP_USERNAME=your_namecheap_username_here
NAMECHEAP_CLIENT_IP=your_server_ip_here  # Optional

# Hostinger Production Account
HOSTINGER_API_KEY=your_hostinger_api_key_here
HOSTINGER_CLIENT_ID=your_hostinger_client_id_here  # Optional
```

### Predefined Credential References

The system includes predefined credential references:

| Provider | Reference | Environment Variables |
|----------|-----------|----------------------|
| GoDaddy | `GODADDY_DEFAULT` | `GODADDY_API_KEY`, `GODADDY_API_SECRET` |
| GoDaddy | `GODADDY_STAGING` | `GODADDY_STAGING_API_KEY`, `GODADDY_STAGING_API_SECRET` |
| Namecheap | `NAMECHEAP_DEFAULT` | `NAMECHEAP_API_KEY`, `NAMECHEAP_USERNAME`, `NAMECHEAP_CLIENT_IP` |
| Namecheap | `NAMECHEAP_STAGING` | `NAMECHEAP_STAGING_API_KEY`, `NAMECHEAP_STAGING_USERNAME` |
| Hostinger | `HOSTINGER_DEFAULT` | `HOSTINGER_API_KEY`, `HOSTINGER_CLIENT_ID` |
| Hostinger | `HOSTINGER_STAGING` | `HOSTINGER_STAGING_API_KEY`, `HOSTINGER_STAGING_CLIENT_ID` |

## Database Schema

The new `secure_provider_credentials` table stores only metadata:

```sql
CREATE TABLE secure_provider_credentials (
    id UUID PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,           -- Provider type
    name VARCHAR(100) NOT NULL,              -- User-friendly name
    account_name VARCHAR(255) NOT NULL,      -- Account identifier  
    credential_reference VARCHAR(100) NOT NULL, -- Reference to env vars
    enabled BOOLEAN NOT NULL DEFAULT true,
    connection_status VARCHAR(50) NOT NULL,
    last_sync TIMESTAMPTZ,
    last_sync_error TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

## Usage in Admin Interface

### Connecting a Provider

1. Go to **Providers** section in the admin interface
2. Click **Connect Provider**
3. Select your provider (GoDaddy, Namecheap, Hostinger)
4. Choose a **Credential Reference** from available options
5. Enter a friendly name and account identifier
6. Test the connection (optional but recommended)
7. Save the connection

### Available Credential References

The admin interface will show available credential references based on which environment variables are set:

- ✅ **Available**: All required environment variables are set
- ❌ **Unavailable**: Some required environment variables are missing

### Example Connection Process

1. **Set Environment Variables**:
   ```bash
   # In your .env file
   GODADDY_API_KEY=3mM44UdWyeo_46cc991d7d9bcc9a_46cc991d7d9bcc9a
   GODADDY_API_SECRET=46cc991d7d9bcc9a46cc991d7d9bcc9a
   ```

2. **Connect in Admin Interface**:
   - Provider: `GoDaddy`
   - Credential Reference: `GODADDY_DEFAULT`
   - Name: `GoDaddy Production Account`
   - Account Name: `production@example.com`

3. **Test Connection**: The system will resolve the environment variables and test the API connection

4. **Save**: Only the reference (`GODADDY_DEFAULT`) is stored in the database

## Migration from Legacy System

### Automatic Migration

The system supports both legacy and secure credentials simultaneously:

1. **Legacy credentials** continue to work (stored in `provider_credentials` table)
2. **New connections** use secure credentials (stored in `secure_provider_credentials` table)
3. **Gradual migration** allows moving to secure credentials over time

### Migration Steps

1. **Set up environment variables** for your providers
2. **Create new secure connections** in the admin interface
3. **Test the new connections** thoroughly
4. **Disable legacy connections** once secure connections are verified
5. **Remove legacy connections** when ready

## API Endpoints

### Get Available Credential Options
```http
GET /api/v1/admin/providers/credentials/options?provider=godaddy
```

Response:
```json
{
  "options": [
    {
      "reference": "GODADDY_DEFAULT",
      "display_name": "GoDaddy Production Account",
      "provider": "godaddy",
      "available": true
    },
    {
      "reference": "GODADDY_STAGING", 
      "display_name": "GoDaddy Staging Account",
      "provider": "godaddy",
      "available": false
    }
  ]
}
```

### Create Secure Connection
```http
POST /api/v1/admin/providers/secure/connect
```

Request:
```json
{
  "provider": "godaddy",
  "name": "GoDaddy Production",
  "account_name": "production@example.com",
  "credential_reference": "GODADDY_DEFAULT",
  "test_connection": true,
  "auto_sync": true,
  "sync_interval_hours": 24
}
```

### Test Secure Connection
```http
POST /api/v1/admin/providers/secure/test
```

Request:
```json
{
  "provider": "godaddy",
  "credential_reference": "GODADDY_DEFAULT"
}
```

## Best Practices

### Environment Variable Management

1. **Never commit `.env` files** to version control
2. **Use different `.env` files** for different environments
3. **Set proper file permissions** on `.env` files (e.g., `chmod 600 .env`)
4. **Use a secrets manager** in production (AWS Secrets Manager, HashiCorp Vault, etc.)

### Credential Security

1. **Rotate credentials regularly** (every 90 days recommended)
2. **Use least-privilege API keys** with minimal required permissions
3. **Monitor API key usage** through provider dashboards
4. **Revoke unused credentials** immediately

### Production Deployment

1. **Use environment-specific credentials** for each deployment
2. **Implement proper secrets management** (not plain text files)
3. **Audit credential access** regularly
4. **Have a credential rotation plan** in place

## Troubleshooting

### Common Issues

1. **"Credential reference not found"**
   - Check that environment variables are set correctly
   - Verify variable names match the predefined references
   - Restart the application after setting new environment variables

2. **"Connection test failed"**
   - Verify API keys are correct and not expired
   - Check network connectivity to provider APIs
   - Ensure API keys have necessary permissions

3. **"Required environment variable not set"**
   - Check the `.env` file for missing variables
   - Verify variable names are spelled correctly
   - Restart the application after adding variables

### Debug Mode

Enable debug logging to troubleshoot credential issues:

```bash
LOG_LEVEL=debug DATABASE_URL=your_db_url go run cmd/server/main.go
```

This will show detailed information about:
- Environment variable loading
- Credential resolution
- API connection attempts
- Provider client creation

## Future Enhancements

### Planned Features

1. **Credential encryption** at rest using application-level encryption
2. **Integration with secrets managers** (AWS Secrets Manager, Vault, etc.)
3. **Credential expiration monitoring** with automatic alerts
4. **Audit logging** for all credential operations
5. **Role-based access control** for credential management

### Extensibility

The secure credential system is designed to be extensible:

1. **Add new providers** by updating the `CredentialReferenceMap`
2. **Support custom credential references** for specific deployments
3. **Integrate with external secret stores** through the `GetenvSecure` function
4. **Add credential validation** and policy enforcement

## Migration Timeline

### Phase 1: Parallel Operation (Current)
- Legacy and secure systems running simultaneously
- New connections use secure credentials
- Existing connections continue with legacy system

### Phase 2: Migration Encouraged (Future)
- Admin interface promotes secure credentials
- Migration tools provided for existing connections
- Legacy system marked as deprecated

### Phase 3: Secure Only (Future)
- Legacy credential system removed
- All connections use secure credentials
- Database cleanup migration provided

This secure credential management system provides a foundation for enterprise-grade security while maintaining ease of use and operational flexibility.
