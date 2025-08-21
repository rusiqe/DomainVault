package providers

import (
	"time"

	"github.com/google/uuid"
	"github.com/rusiqe/domainvault/internal/types"
)

// MockClient implements RegistrarClient for testing
type MockClient struct {
	name    string
	domains []types.Domain
}

// NewMockClient creates a mock provider client
func NewMockClient(creds ProviderCredentials) (*MockClient, error) {
	// Generate some mock domains
	now := time.Now()
	domains := []types.Domain{
		{
			ID:        uuid.New().String(),
			Name:      "example.com",
			Provider:  "mock",
			ExpiresAt: now.AddDate(1, 0, 0), // 1 year from now
			CreatedAt: now.AddDate(-1, 0, 0), // 1 year ago
			UpdatedAt: now,
		},
		{
			ID:        uuid.New().String(),
			Name:      "test.org",
			Provider:  "mock",
			ExpiresAt: now.AddDate(0, 1, 0), // 1 month from now
			CreatedAt: now.AddDate(-2, 0, 0), // 2 years ago
			UpdatedAt: now,
		},
		{
			ID:        uuid.New().String(),
			Name:      "demo.net",
			Provider:  "mock",
			ExpiresAt: now.AddDate(0, 0, 30), // 30 days from now
			CreatedAt: now.AddDate(0, -6, 0), // 6 months ago
			UpdatedAt: now,
		},
	}

	return &MockClient{
		name:    "mock",
		domains: domains,
	}, nil
}

// FetchDomains returns mock domain data
func (m *MockClient) FetchDomains() ([]types.Domain, error) {
	// Simulate API delay
	time.Sleep(100 * time.Millisecond)
	
	// Return a copy of domains to prevent modification
	result := make([]types.Domain, len(m.domains))
	copy(result, m.domains)
	
	return result, nil
}

// GetProviderName returns the provider name
func (m *MockClient) GetProviderName() string {
	return m.name
}

// FetchDNSRecords returns mock DNS records for a domain
func (m *MockClient) FetchDNSRecords(domain string) ([]types.DNSRecord, error) {
	// Simulate API delay
	time.Sleep(50 * time.Millisecond)
	
	now := time.Now()
	
	// Generate mock DNS records based on domain name
	var records []types.DNSRecord
	
	// Basic A records
	records = append(records, types.DNSRecord{
		ID:        uuid.New().String(),
		Type:      "A",
		Name:      "@",
		Value:     "192.0.2.1",
		TTL:       3600,
		CreatedAt: now,
		UpdatedAt: now,
	})
	
	records = append(records, types.DNSRecord{
		ID:        uuid.New().String(),
		Type:      "A",
		Name:      "www",
		Value:     "192.0.2.1",
		TTL:       3600,
		CreatedAt: now,
		UpdatedAt: now,
	})
	
	// MX record
	priority := 10
	records = append(records, types.DNSRecord{
		ID:        uuid.New().String(),
		Type:      "MX",
		Name:      "@",
		Value:     "mail." + domain,
		TTL:       3600,
		Priority:  &priority,
		CreatedAt: now,
		UpdatedAt: now,
	})
	
	// TXT record for SPF
	records = append(records, types.DNSRecord{
		ID:        uuid.New().String(),
		Type:      "TXT",
		Name:      "@",
		Value:     "v=spf1 include:_spf." + domain + " ~all",
		TTL:       3600,
		CreatedAt: now,
		UpdatedAt: now,
	})
	
	// CNAME for subdomains
	if domain == "example.com" {
		records = append(records, types.DNSRecord{
			ID:        uuid.New().String(),
			Type:      "CNAME",
			Name:      "api",
			Value:     "api-server.example.com",
			TTL:       1800,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	
	return records, nil
}

// AddMockDomain adds a domain to the mock provider (for testing)
func (m *MockClient) AddMockDomain(domain types.Domain) {
	domain.Provider = m.name
	if domain.ID == "" {
		domain.ID = uuid.New().String()
	}
	m.domains = append(m.domains, domain)
}
