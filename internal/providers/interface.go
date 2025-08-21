package providers

import (
	"github.com/rusiqe/domainvault/internal/types"
)

// RegistrarClient defines the interface for domain registrar integrations
type RegistrarClient interface {
	// Core operations
	FetchDomains() ([]types.Domain, error)
	GetProviderName() string
	
	// DNS operations
	FetchDNSRecords(domain string) ([]types.DNSRecord, error)
	
	// Future hooks for MVP expansion
	// RenewDomain(domainID string) error
	// UpdateDNS(domain string, records []types.DNSRecord) error
	// GetDomainInfo(domain string) (*types.Domain, error)
}

// ProviderCredentials holds authentication data for providers
type ProviderCredentials map[string]interface{}

// ClientFactory creates registrar clients
func NewClient(provider string, creds ProviderCredentials) (RegistrarClient, error) {
	switch provider {
	case "godaddy":
		return NewGoDaddyClient(creds)
	case "namecheap":
		return NewNamecheapClient(creds)
	case "hostinger":
		return NewHostingerClient(creds)
	case "cloudflare":
		return NewCloudflareClient(creds)
	case "mock":
		return NewMockClient(creds)
	default:
		return nil, types.ErrUnsupportedProvider
	}
}

// ValidateCredentials checks if provider credentials are valid
func ValidateCredentials(provider string, creds ProviderCredentials) error {
	switch provider {
	case "godaddy":
		if _, ok := creds["api_key"]; !ok {
			return types.ErrMissingConfig
		}
		if _, ok := creds["api_secret"]; !ok {
			return types.ErrMissingConfig
		}
	case "namecheap":
		if _, ok := creds["api_key"]; !ok {
			return types.ErrMissingConfig
		}
		if _, ok := creds["username"]; !ok {
			return types.ErrMissingConfig
		}
	case "hostinger":
		if _, ok := creds["api_key"]; !ok {
			return types.ErrMissingConfig
		}
	case "cloudflare":
		if _, ok := creds["api_token"]; !ok {
			return types.ErrMissingConfig
		}
	case "mock":
		// Mock provider doesn't require credentials
		return nil
	default:
		return types.ErrUnsupportedProvider
	}
	return nil
}
