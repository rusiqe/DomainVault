package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rusiqe/domainvault/internal/types"
)

// GoDaddyClient implements RegistrarClient for GoDaddy API
type GoDaddyClient struct {
	apiKey    string
	apiSecret string
	baseURL   string
	client    *http.Client
}

// GoDaddyDomain represents domain data from GoDaddy API
type GoDaddyDomain struct {
	Domain    string    `json:"domain"`
	DomainId  int64     `json:"domainId"`
	ExpiresAt time.Time `json:"expires"`
	CreatedAt time.Time `json:"createdAt"`
	Renewable bool      `json:"renewable"`
	Status    string    `json:"status"`
}

// NewGoDaddyClient creates a new GoDaddy client
func NewGoDaddyClient(creds ProviderCredentials) (*GoDaddyClient, error) {
	apiKey, ok := creds["api_key"].(string)
	if !ok {
		return nil, types.ErrMissingConfig
	}
	
	apiSecret, ok := creds["api_secret"].(string)
	if !ok {
		return nil, types.ErrMissingConfig
	}

	return &GoDaddyClient{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   "https://api.godaddy.com/v1",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// FetchDomains retrieves domains from GoDaddy API
func (g *GoDaddyClient) FetchDomains() ([]types.Domain, error) {
	url := fmt.Sprintf("%s/domains", g.baseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", g.apiKey, g.apiSecret))
	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
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

	var godaddyDomains []GoDaddyDomain
	if err := json.NewDecoder(resp.Body).Decode(&godaddyDomains); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to internal domain format
	domains := make([]types.Domain, len(godaddyDomains))
	for i, gd := range godaddyDomains {
		domains[i] = types.Domain{
			ID:        uuid.New().String(), // Generate new UUID
			Name:      gd.Domain,
			Provider:  "godaddy",
			ExpiresAt: gd.ExpiresAt,
			CreatedAt: gd.CreatedAt,
			UpdatedAt: time.Now(),
		}
	}

	return domains, nil
}

// GetProviderName returns the provider name
func (g *GoDaddyClient) GetProviderName() string {
	return "godaddy"
}

// GoDaddyDNSRecord represents DNS record data from GoDaddy API
type GoDaddyDNSRecord struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Data     string `json:"data"`
	TTL      int    `json:"ttl"`
	Priority *int   `json:"priority,omitempty"`
	Weight   *int   `json:"weight,omitempty"`
	Port     *int   `json:"port,omitempty"`
}

// FetchDNSRecords retrieves DNS records for a domain from GoDaddy API
func (g *GoDaddyClient) FetchDNSRecords(domain string) ([]types.DNSRecord, error) {
	url := fmt.Sprintf("%s/domains/%s/records", g.baseURL, domain)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS request: %w", err)
	}

	// Set authorization header
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", g.apiKey, g.apiSecret))
	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DNS records: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, types.ErrProviderAuth
	}
	
	if resp.StatusCode == 429 {
		return nil, types.ErrProviderRateLimit
	}

	if resp.StatusCode == 404 {
		// Domain not found or no DNS records
		return []types.DNSRecord{}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code for DNS records: %d", resp.StatusCode)
	}

	var godaddyRecords []GoDaddyDNSRecord
	if err := json.NewDecoder(resp.Body).Decode(&godaddyRecords); err != nil {
		return nil, fmt.Errorf("failed to decode DNS response: %w", err)
	}

	// Convert to internal DNS record format
	dnsRecords := make([]types.DNSRecord, 0, len(godaddyRecords))
	for _, gr := range godaddyRecords {
		dnsRecord := types.DNSRecord{
			ID:        uuid.New().String(), // Generate new UUID
			Type:      gr.Type,
			Name:      gr.Name,
			Value:     gr.Data,
			TTL:       gr.TTL,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		// Handle additional fields for specific record types
		if gr.Priority != nil {
			dnsRecord.Priority = gr.Priority
		}
		if gr.Weight != nil {
			dnsRecord.Weight = gr.Weight
		}
		if gr.Port != nil {
			dnsRecord.Port = gr.Port
		}
		
		dnsRecords = append(dnsRecords, dnsRecord)
	}

	return dnsRecords, nil
}

// Future implementations for MVP expansion:
// func (g *GoDaddyClient) RenewDomain(domainID string) error { ... }
// func (g *GoDaddyClient) UpdateDNS(domain string, records []types.DNSRecord) error { ... }
