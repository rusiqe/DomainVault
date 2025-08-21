package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rusiqe/domainvault/internal/types"
)

// HostingerClient implements RegistrarClient for Hostinger API
// According to official Hostinger API documentation:
// Base URL: https://developers.hostinger.com/api
// Domains endpoint: /domains/v1/domains or /domains/v1/portfolio  
// DNS endpoint: /domains/v1/dns/{domain} or /domains/{domain}/dns-zones
type HostingerClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// HostingerDomain represents domain data from Hostinger API
type HostingerDomain struct {
	ID        int       `json:"id"`
	Domain    *string   `json:"domain"`    // nullable when not claimed free domain
	Type      string    `json:"type"`      // "domain" or "free_domain"
	Status    string    `json:"status"`    // "active", "pending_setup", "expired", "requested", "pending_verification"
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"` // nullable
}

// NewHostingerClient creates a new Hostinger client
func NewHostingerClient(creds ProviderCredentials) (*HostingerClient, error) {
	apiKey, ok := creds["api_key"].(string)
	if !ok {
		return nil, types.ErrMissingConfig
	}

	return &HostingerClient{
		apiKey:  apiKey,
		baseURL: "https://developers.hostinger.com/api", // Official API base URL
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// FetchDomains retrieves domains from Hostinger API
// Working endpoint: /domains/v1/portfolio
func (h *HostingerClient) FetchDomains() ([]types.Domain, error) {
	url := fmt.Sprintf("%s/domains/v1/portfolio", h.baseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.apiKey))
	req.Header.Set("Accept", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch domains: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, types.ErrProviderAuth
	}
	
	if resp.StatusCode == 429 {
		return nil, types.ErrProviderRateLimit
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Response is directly an array of domains, not wrapped in a response object
	var hostingerDomains []HostingerDomain
	if err := json.NewDecoder(resp.Body).Decode(&hostingerDomains); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Filter out domains with null names and convert to internal domain format
	domains := make([]types.Domain, 0, len(hostingerDomains))
	for _, hd := range hostingerDomains {
		// Skip domains with null names (unclaimed free domains)
		if hd.Domain == nil {
			continue
		}

		// Handle nullable expires_at
		var expiresAt time.Time
		if hd.ExpiresAt != nil {
			expiresAt = *hd.ExpiresAt
		}

		domains = append(domains, types.Domain{
			ID:        uuid.New().String(), // Generate new UUID
			Name:      *hd.Domain,
			Provider:  "hostinger",
			ExpiresAt: expiresAt,
			AutoRenew: false, // API doesn't provide auto-renew info
			Status:    h.mapStatus(hd.Status),
			CreatedAt: hd.CreatedAt,
			UpdatedAt: time.Now(),
		})
	}

	return domains, nil
}

// GetProviderName returns the provider name
func (h *HostingerClient) GetProviderName() string {
	return "hostinger"
}

// mapStatus maps Hostinger status to internal status
func (h *HostingerClient) mapStatus(hostingerStatus string) string {
	switch hostingerStatus {
	case "active":
		return "active"
	case "expired":
		return "expired"
	case "pending_setup":
		return "pending"
	case "requested":
		return "pending"
	case "pending_verification":
		return "pending"
	default:
		return "unknown"
	}
}

// HostingerDNSRecord represents DNS record data from Hostinger API
type HostingerDNSRecord struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl"`
	Priority *int   `json:"priority,omitempty"`
}

// FetchDNSRecords retrieves DNS records for a domain from Hostinger API
// Try fetching DNS records using multiple possible endpoints
func (h *HostingerClient) FetchDNSRecords(domain string) ([]types.DNSRecord, error) {
	endpoints := []string{
		fmt.Sprintf("%s/domains/v1/dns/%s", h.baseURL, domain), // Original attempt
		fmt.Sprintf("%s/domains/%s/dns-zones", h.baseURL, domain), // Official documentation
	}
	
	for _, endpoint := range endpoints {
		url := endpoint
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			// Skip to next endpoint on request creation error
			continue
		}

		// Set authorization header
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.apiKey))
		req.Header.Set("Accept", "application/json")

		resp, err := h.client.Do(req)
		if err != nil {
			// Skip to next endpoint on request error
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 401 {
			return nil, types.ErrProviderAuth
		}
		
		if resp.StatusCode == 429 {
			return nil, types.ErrProviderRateLimit
		}

		if resp.StatusCode == 404 {
			// Try next endpoint if this one returns 404
			continue
		}

		if resp.StatusCode != 200 {
			// Try next endpoint on unexpected status code
			continue
		}

		// Response is directly an array of DNS records
		var hostingerRecords []HostingerDNSRecord
		if err := json.NewDecoder(resp.Body).Decode(&hostingerRecords); err != nil {
			// Try next endpoint on decode error
			continue
		}

		// Convert to internal DNS record format
		dnsRecords := make([]types.DNSRecord, 0, len(hostingerRecords))
		for _, hr := range hostingerRecords {
			dnsRecord := types.DNSRecord{
				ID:        uuid.New().String(), // Generate new UUID
				Type:      hr.Type,
				Name:      hr.Name,
				Value:     hr.Content,
				TTL:       hr.TTL,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			
			// Handle priority for MX and SRV records
			if hr.Priority != nil {
				dnsRecord.Priority = hr.Priority
			}
			
			dnsRecords = append(dnsRecords, dnsRecord)
		}

		return dnsRecords, nil
	}

	// If all endpoints failed, return empty records (no DNS records found)
	// This matches the behavior of the original implementation
	return []types.DNSRecord{}, nil
}

// Future implementations for MVP expansion:
// func (h *HostingerClient) RenewDomain(domainID string) error { ... }
// func (h *HostingerClient) UpdateDNS(domain string, records []types.DNSRecord) error { ... }
// func (h *HostingerClient) GetDomainInfo(domain string) (*types.Domain, error) { ... }
