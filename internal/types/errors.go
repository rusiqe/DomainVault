package types

import "errors"

// Domain validation errors
var (
	ErrInvalidDomainName = errors.New("invalid domain name")
	ErrInvalidProvider   = errors.New("invalid provider")
	ErrDomainNotFound    = errors.New("domain not found")
	ErrDomainExists      = errors.New("domain already exists")
)

// Provider errors
var (
	ErrUnsupportedProvider = errors.New("unsupported provider")
	ErrProviderAuth       = errors.New("provider authentication failed")
	ErrProviderRateLimit  = errors.New("provider rate limit exceeded")
	ErrProviderTimeout    = errors.New("provider request timeout")
)

// Database errors
var (
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrDatabaseMigration  = errors.New("database migration failed")
)

// Configuration errors
var (
	ErrInvalidConfig = errors.New("invalid configuration")
	ErrMissingConfig = errors.New("missing required configuration")
)
