package types

import (
	"time"
)

// Domain represents a domain name with its metadata
type Domain struct {
	ID          string    `json:"id" db:"id"`                   // UUIDv7
	Name        string    `json:"name" db:"name"`               // FQDN
	Provider    string    `json:"provider" db:"provider"`       // Registrar name
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`   // Expiration date
	CreatedAt   time.Time `json:"created_at" db:"created_at"`   // Record creation
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`   // Last update
	CategoryID  *string   `json:"category_id,omitempty" db:"category_id"` // Category assignment
	ProjectID   *string   `json:"project_id,omitempty" db:"project_id"`   // Project assignment
	AutoRenew   bool      `json:"auto_renew" db:"auto_renew"`              // Auto-renewal setting
	RenewalPrice *float64 `json:"renewal_price,omitempty" db:"renewal_price"` // Annual renewal cost
	Status      string    `json:"status" db:"status"`                      // active, expired, transferred, etc.
	Tags        []string  `json:"tags,omitempty" db:"tags"`                // Organization tags
	
	// HTTP Status monitoring
	HTTPStatus      *int       `json:"http_status,omitempty" db:"http_status"`           // Last HTTP status code
	LastStatusCheck *time.Time `json:"last_status_check,omitempty" db:"last_status_check"` // When status was last checked
	StatusMessage   *string    `json:"status_message,omitempty" db:"status_message"`     // Human-readable status message
	
	// DNS Records (populated on demand)
	DNSRecords []DNSRecord `json:"dns_records,omitempty" db:"-"`

	// Future expansion hooks (commented out for now)
	// Nameservers  []string    `json:"nameservers,omitempty" db:"nameservers"`     // Custom nameservers
	// RegistrantInfo map[string]string `json:"registrant_info,omitempty" db:"registrant_info"` // Whois data
}


// DomainSummary provides aggregated domain statistics
type DomainSummary struct {
	Total       int                    `json:"total"`
	ByProvider  map[string]int         `json:"by_provider"`
	ExpiringIn  map[string]int         `json:"expiring_in"` // "30_days", "90_days", etc.
	LastSync    time.Time              `json:"last_sync"`
}

// DomainFilter for querying domains
type DomainFilter struct {
	Provider     string    `json:"provider,omitempty"`
	ExpiresAfter *time.Time `json:"expires_after,omitempty"`
	ExpiresBefore *time.Time `json:"expires_before,omitempty"`
	Search       string    `json:"search,omitempty"` // Search in domain name
	CategoryID   *string   `json:"category_id,omitempty"`
	ProjectID    *string   `json:"project_id,omitempty"`
	Limit        int       `json:"limit,omitempty"`
	Offset       int       `json:"offset,omitempty"`
}

// Category represents a domain categorization
type Category struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Color       string    `json:"color" db:"color"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Project represents a domain project grouping
type Project struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Color       string    `json:"color" db:"color"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ProviderCredentials stores encrypted API credentials for domain providers
type ProviderCredentials struct {
	ID              string            `json:"id" db:"id"`
	Provider        string            `json:"provider" db:"provider"`
	Name            string            `json:"name" db:"name"` // User-friendly name
	Credentials     map[string]string `json:"credentials" db:"credentials"` // Encrypted credentials
	Enabled         bool              `json:"enabled" db:"enabled"`
	LastSync        *time.Time        `json:"last_sync,omitempty" db:"last_sync"`
	LastSyncError   *string           `json:"last_sync_error,omitempty" db:"last_sync_error"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
}

// ImportRequest represents a manual domain import request
type ImportRequest struct {
	Provider        string            `json:"provider"`
	Credentials     map[string]string `json:"credentials"`
	StoreCredentials bool             `json:"store_credentials"`
	CredentialsName string            `json:"credentials_name,omitempty"`
	CategoryID      *string           `json:"category_id,omitempty"`
	ProjectID       *string           `json:"project_id,omitempty"`
}

// User represents an admin user
type User struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never expose in JSON
	Role         string    `json:"role" db:"role"`
	Enabled      bool      `json:"enabled" db:"enabled"`
	LastLogin    *time.Time `json:"last_login,omitempty" db:"last_login"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// LoginRequest represents a login attempt
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a successful login
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User   `json:"user"`
}

// DNSRecord represents a DNS record with full details
type DNSRecord struct {
	ID       string `json:"id" db:"id"`
	DomainID string `json:"domain_id" db:"domain_id"`
	Type     string `json:"type" db:"type"`         // A, AAAA, CNAME, MX, TXT, NS, etc.
	Name     string `json:"name" db:"name"`         // Subdomain or @ for root
	Value    string `json:"value" db:"value"`       // IP, hostname, text content
	TTL      int    `json:"ttl" db:"ttl"`           // Time to live in seconds
	Priority *int   `json:"priority,omitempty" db:"priority"` // For MX records
	Weight   *int   `json:"weight,omitempty" db:"weight"`     // For SRV records
	Port     *int   `json:"port,omitempty" db:"port"`         // For SRV records
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DomainPurchaseRequest represents a bulk domain purchase request
type DomainPurchaseRequest struct {
	Domains     []string `json:"domains" binding:"required"`
	Provider    string   `json:"provider" binding:"required"`
	CredentialsID string `json:"credentials_id" binding:"required"`
	CategoryID  *string  `json:"category_id,omitempty"`
	ProjectID   *string  `json:"project_id,omitempty"`
	Years       int      `json:"years" binding:"min=1,max=10"`
	AutoRenew   bool     `json:"auto_renew"`
}

// DomainDecommissionRequest represents a bulk domain decommission request
type DomainDecommissionRequest struct {
	DomainIDs []string `json:"domain_ids" binding:"required"`
	StopAutoRenew bool `json:"stop_auto_renew"`
	TransferOut   bool `json:"transfer_out"`
	DeleteDNS     bool `json:"delete_dns"`
}

// BulkSyncRequest represents a manual bulk sync request
type BulkSyncRequest struct {
	Providers     []string `json:"providers,omitempty"`
	CredentialsIDs []string `json:"credentials_ids,omitempty"`
	ForceRefresh  bool     `json:"force_refresh"`
}

// Validate checks if domain data is valid
func (d *Domain) Validate() error {
	if d.Name == "" {
		return ErrInvalidDomainName
	}
	if d.Provider == "" {
		return ErrInvalidProvider
	}
	return nil
}

// IsExpiringSoon checks if domain expires within the given duration
func (d *Domain) IsExpiringSoon(duration time.Duration) bool {
	return time.Until(d.ExpiresAt) <= duration
}

// DaysUntilExpiration returns days until domain expiration
func (d *Domain) DaysUntilExpiration() int {
	if d.ExpiresAt.IsZero() {
		return -1 // Unknown expiration
	}
	duration := time.Until(d.ExpiresAt)
	return int(duration.Hours() / 24)
}

// Validate checks if category data is valid
func (c *Category) Validate() error {
	if c.Name == "" {
		return ErrInvalidDomainName // Reuse existing error for now
	}
	return nil
}

// Validate checks if project data is valid
func (p *Project) Validate() error {
	if p.Name == "" {
		return ErrInvalidDomainName // Reuse existing error for now
	}
	return nil
}

// Validate checks if provider credentials are valid
func (pc *ProviderCredentials) Validate() error {
	if pc.Provider == "" {
		return ErrInvalidProvider
	}
	if pc.Name == "" {
		return ErrInvalidDomainName // Reuse existing error for now
	}
	if len(pc.Credentials) == 0 {
		return ErrInvalidProvider
	}
	return nil
}
