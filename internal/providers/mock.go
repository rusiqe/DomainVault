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

// AddMockDomain adds a domain to the mock provider (for testing)
func (m *MockClient) AddMockDomain(domain types.Domain) {
	domain.Provider = m.name
	if domain.ID == "" {
		domain.ID = uuid.New().String()
	}
	m.domains = append(m.domains, domain)
}
