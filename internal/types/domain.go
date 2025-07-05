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

	// Future expansion hooks (commented out for now)
	// RenewalPrice float64     `json:"renewal_price,omitempty" db:"renewal_price"` // Cost monitoring
	// DNSRecords   []DNSRecord `json:"dns_records,omitempty" db:"dns_records"`     // DNS management
	// AutoRenew    bool        `json:"auto_renew,omitempty" db:"auto_renew"`       // Renewal system
	// Tags         []string    `json:"tags,omitempty" db:"tags"`                   // Organization
}

// DNSRecord represents a DNS record (future use)
type DNSRecord struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
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
	Limit        int       `json:"limit,omitempty"`
	Offset       int       `json:"offset,omitempty"`
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
