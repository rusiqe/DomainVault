package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
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
	Tags        TagsSlice `json:"tags,omitempty" db:"tags"`                // Organization tags
	
	// HTTP Status monitoring
	HTTPStatus      *int       `json:"http_status,omitempty" db:"http_status"`           // Last HTTP status code
	LastStatusCheck *time.Time `json:"last_status_check,omitempty" db:"last_status_check"` // When status was last checked
	StatusMessage   *string    `json:"status_message,omitempty" db:"status_message"`     // Human-readable status message
	
	// UptimeRobot monitoring
	UptimeRobotMonitorID *int     `json:"uptime_robot_monitor_id,omitempty" db:"uptime_robot_monitor_id"` // UptimeRobot monitor ID
	UptimeRatio          *float64 `json:"uptime_ratio,omitempty" db:"uptime_ratio"`                       // Uptime percentage (0-100)
	ResponseTime         *int     `json:"response_time,omitempty" db:"response_time"`                     // Average response time in ms
	MonitorStatus        *string  `json:"monitor_status,omitempty" db:"monitor_status"`                   // up, down, paused, seems_down
	LastDowntime         *time.Time `json:"last_downtime,omitempty" db:"last_downtime"`                   // Last recorded downtime
	
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

// CredentialsMap is a custom type for handling JSON marshaling/unmarshaling of credentials
type CredentialsMap map[string]string

// Value implements the driver.Valuer interface for database storage
func (c CredentialsMap) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface for database retrieval
func (c *CredentialsMap) Scan(value interface{}) error {
	if value == nil {
		*c = make(CredentialsMap)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	default:
		return fmt.Errorf("cannot scan %T into CredentialsMap", value)
	}
}

// TagsSlice is a custom type for handling JSON marshaling/unmarshaling of tags
type TagsSlice []string

// Value implements the driver.Valuer interface for database storage
func (t TagsSlice) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan implements the sql.Scanner interface for database retrieval
func (t *TagsSlice) Scan(value interface{}) error {
	if value == nil {
		*t = make(TagsSlice, 0)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, t)
	case string:
		return json.Unmarshal([]byte(v), t)
	default:
		return fmt.Errorf("cannot scan %T into TagsSlice", value)
	}
}

// ProviderCredentials stores encrypted API credentials for domain providers
type ProviderCredentials struct {
	ID              string         `json:"id" db:"id"`
	Provider        string         `json:"provider" db:"provider"`           // Provider type (godaddy, namecheap)
	Name            string         `json:"name" db:"name"`                   // User-friendly name
	AccountName     string         `json:"account_name" db:"account_name"`   // Account identifier
	Credentials     CredentialsMap `json:"credentials" db:"credentials"`     // API credentials (key, secret, etc.)
	Enabled         bool           `json:"enabled" db:"enabled"`
	ConnectionStatus string        `json:"connection_status" db:"connection_status"` // connected, error, testing
	LastSync        *time.Time     `json:"last_sync,omitempty" db:"last_sync"`
	LastSyncError   *string        `json:"last_sync_error,omitempty" db:"last_sync_error"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

// ProviderInfo represents information about supported providers
type ProviderInfo struct {
	Name         string                    `json:"name"`          // Provider name (godaddy, namecheap)
	DisplayName  string                    `json:"display_name"`  // Human-readable name
	Description  string                    `json:"description"`   // Provider description
	Fields       []ProviderFieldInfo      `json:"fields"`        // Required credential fields
	DocumentationURL string               `json:"documentation_url,omitempty"` // Setup guide URL
}

// ProviderFieldInfo describes a credential field
type ProviderFieldInfo struct {
	Name         string `json:"name"`         // Field name (api_key, api_secret)
	DisplayName  string `json:"display_name"` // Human-readable name
	Type         string `json:"type"`         // text, password, email
	Required     bool   `json:"required"`     // Whether field is required
	Description  string `json:"description"`  // Field description
	Placeholder  string `json:"placeholder,omitempty"` // Example value
}

// ProviderConnectionRequest represents a provider connection request
type ProviderConnectionRequest struct {
	Provider          string            `json:"provider" binding:"required"`
	Name              string            `json:"name" binding:"required"`         // User-friendly name
	AccountName       string            `json:"account_name" binding:"required"` // Account identifier
	Credentials       map[string]string `json:"credentials" binding:"required"`  // API credentials
	TestConnection    bool              `json:"test_connection"`                 // Test before saving
	AutoSync          bool              `json:"auto_sync"`                       // Run initial sync if test passes
	SyncIntervalHours int               `json:"sync_interval_hours"`             // Auto-sync interval in hours
}

// ProviderConnectionResponse represents the result of a connection attempt
type ProviderConnectionResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	ProviderID   string `json:"provider_id,omitempty"`   // ID if successful
	DomainsFound int    `json:"domains_found,omitempty"` // Number of domains found during test
	SyncStarted  bool   `json:"sync_started,omitempty"`  // Whether initial sync was started
}

// DomainSearchRequest represents a domain availability search request
type DomainSearchRequest struct {
	Domains    []string `json:"domains" binding:"required"`
	CouponCode string   `json:"coupon_code,omitempty"`
}

// DomainSearchResult represents the result of a domain search
type DomainSearchResult struct {
	Domain     string                 `json:"domain"`
	Available  bool                   `json:"available"`
	Premium    bool                   `json:"premium,omitempty"`
	Pricing    DomainPricing          `json:"pricing,omitempty"`
	Providers  []DomainProviderOption `json:"providers,omitempty"`
	Message    string                 `json:"message,omitempty"`
}

// DomainPricing represents pricing information for a domain
type DomainPricing struct {
	PurchasePrice float64 `json:"purchase_price"`
	RenewalPrice  float64 `json:"renewal_price"`
	Currency      string  `json:"currency"`
	Period        int     `json:"period"`        // Years
	CouponDiscount float64 `json:"coupon_discount,omitempty"`
}

// DomainProviderOption represents a provider option for domain purchase
type DomainProviderOption struct {
	ProviderID   string        `json:"provider_id"`
	ProviderName string        `json:"provider_name"`
	DisplayName  string        `json:"display_name"`
	Pricing      DomainPricing `json:"pricing"`
	Supported    bool          `json:"supported"`    // Whether we have credentials for this provider
	Recommended  bool          `json:"recommended"`  // Whether this is the recommended option
}

// DomainPurchaseRequest represents a domain purchase request
type DomainPurchaseRequest struct {
	Domains     []DomainPurchaseItem `json:"domains" binding:"required"`
	ProviderID  string               `json:"provider_id" binding:"required"`
	CouponCode  string               `json:"coupon_code,omitempty"`
	AutoRenew   bool                 `json:"auto_renew"`
	CategoryID  *string              `json:"category_id,omitempty"`
	ProjectID   *string              `json:"project_id,omitempty"`
}

// DomainPurchaseItem represents a single domain to purchase
type DomainPurchaseItem struct {
	Domain string `json:"domain" binding:"required"`
	Period int    `json:"period"`  // Years to register
}

// DomainPurchaseResponse represents the result of a domain purchase
type DomainPurchaseResponse struct {
	Success        bool                    `json:"success"`
	Message        string                  `json:"message"`
	PurchasedDomains []PurchasedDomainInfo `json:"purchased_domains,omitempty"`
	FailedDomains    []FailedDomainPurchase `json:"failed_domains,omitempty"`
	TotalCost      float64                 `json:"total_cost,omitempty"`
	Currency       string                  `json:"currency,omitempty"`
	TransactionID  string                  `json:"transaction_id,omitempty"`
}

// PurchasedDomainInfo represents information about a successfully purchased domain
type PurchasedDomainInfo struct {
	Domain      string    `json:"domain"`
	DomainID    string    `json:"domain_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	Cost        float64   `json:"cost"`
	Period      int       `json:"period"`
}

// FailedDomainPurchase represents information about a failed domain purchase
type FailedDomainPurchase struct {
	Domain string `json:"domain"`
	Error  string `json:"error"`
}

// WebsiteStatusRequest represents a website status check request
type WebsiteStatusRequest struct {
	Domains []string `json:"domains" binding:"required"`
}

// WebsiteStatusResult represents the result of a website status check
type WebsiteStatusResult struct {
	Domain           string    `json:"domain"`
	HTTPStatus       int       `json:"http_status"`
	StatusMessage    string    `json:"status_message"`
	ResponseTime     int64     `json:"response_time_ms"`
	SSLStatus        string    `json:"ssl_status,omitempty"`
	RedirectURL      string    `json:"redirect_url,omitempty"`
	LastChecked      time.Time `json:"last_checked"`
	Error            string    `json:"error,omitempty"`
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

// UptimeRobotConfig represents UptimeRobot configuration
type UptimeRobotConfig struct {
	APIKey       string   `json:"api_key" binding:"required"`
	Enabled      bool     `json:"enabled"`
	Interval     int      `json:"interval"`           // Check interval in seconds (60, 120, 300, 600, 900, 1800, 3600)
	AlertContacts []string `json:"alert_contacts"`     // Alert contact IDs
	AutoCreateMonitors bool `json:"auto_create_monitors"` // Automatically create monitors for new domains
}

// UptimeRobotMonitorRequest represents a request to create/update UptimeRobot monitoring
type UptimeRobotMonitorRequest struct {
	DomainIDs     []string `json:"domain_ids" binding:"required"`
	MonitorType   string   `json:"monitor_type"`   // http, ping, keyword
	Interval      int      `json:"interval"`       // Check interval in seconds
	AlertContacts []string `json:"alert_contacts"` // Alert contact IDs
	KeywordCheck  bool     `json:"keyword_check,omitempty"`
	KeywordValue  string   `json:"keyword_value,omitempty"`
	KeywordExists bool     `json:"keyword_exists,omitempty"`
}

// UptimeRobotSyncResponse represents the response from UptimeRobot sync operations
type UptimeRobotSyncResponse struct {
	Success       bool                    `json:"success"`
	Message       string                  `json:"message"`
	MonitorsSync  int                     `json:"monitors_synced"`
	MonitorsCreated int                   `json:"monitors_created"`
	MonitorsUpdated int                   `json:"monitors_updated"`
	MonitorsFailed  int                   `json:"monitors_failed"`
	Results       []DomainMonitorResult   `json:"results,omitempty"`
}

// DomainMonitorResult represents the result of monitoring setup for a domain
type DomainMonitorResult struct {
	DomainID      string  `json:"domain_id"`
	DomainName    string  `json:"domain_name"`
	MonitorID     *int    `json:"monitor_id,omitempty"`
	Action        string  `json:"action"` // created, updated, failed, skipped
	Success       bool    `json:"success"`
	Message       string  `json:"message"`
	UptimeRatio   *float64 `json:"uptime_ratio,omitempty"`
	ResponseTime  *int    `json:"response_time,omitempty"`
	Error         string  `json:"error,omitempty"`
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
