package providers

import (
	"fmt"
	"log"

	"github.com/rusiqe/domainvault/internal/types"
)

// ProviderService manages provider information and connections
type ProviderService struct {
	supportedProviders map[string]types.ProviderInfo
}

// NewProviderService creates a new provider service
func NewProviderService() *ProviderService {
	return &ProviderService{
		supportedProviders: initializeSupportedProviders(),
	}
}

// GetSupportedProviders returns all supported providers
func (ps *ProviderService) GetSupportedProviders() []types.ProviderInfo {
	providers := make([]types.ProviderInfo, 0, len(ps.supportedProviders))
	for _, provider := range ps.supportedProviders {
		providers = append(providers, provider)
	}
	return providers
}

// GetProviderInfo returns information about a specific provider
func (ps *ProviderService) GetProviderInfo(providerName string) (types.ProviderInfo, error) {
	provider, exists := ps.supportedProviders[providerName]
	if !exists {
		return types.ProviderInfo{}, fmt.Errorf("provider %s not supported", providerName)
	}
	return provider, nil
}

// TestConnection tests if the provided credentials work for a provider
func (ps *ProviderService) TestConnection(provider string, credentials map[string]string) (*types.ProviderConnectionResponse, error) {
	// Validate provider is supported
	providerInfo, err := ps.GetProviderInfo(provider)
	if err != nil {
		return &types.ProviderConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Unsupported provider: %s", provider),
		}, nil
	}

	// Validate required fields are present
	for _, field := range providerInfo.Fields {
		if field.Required {
			if _, exists := credentials[field.Name]; !exists {
				return &types.ProviderConnectionResponse{
					Success: false,
					Message: fmt.Sprintf("Missing required field: %s", field.DisplayName),
				}, nil
			}
		}
	}

	// Convert credentials to ProviderCredentials format
	providerCreds := make(ProviderCredentials)
	for key, value := range credentials {
		providerCreds[key] = value
	}

	// Create a test client and try to fetch domains
	client, err := NewClient(provider, providerCreds)
	if err != nil {
		return &types.ProviderConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create client: %v", err),
		}, nil
	}

	// Test the connection by trying to fetch domains
	domains, err := client.FetchDomains()
	if err != nil {
		return &types.ProviderConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Connection test failed: %v", err),
		}, nil
	}

	return &types.ProviderConnectionResponse{
		Success:      true,
		Message:      "Connection successful",
		DomainsFound: len(domains),
	}, nil
}

// IsSupported checks if a provider is supported
func (ps *ProviderService) IsSupported(provider string) bool {
	_, exists := ps.supportedProviders[provider]
	return exists
}

// initializeSupportedProviders returns the map of supported providers
func initializeSupportedProviders() map[string]types.ProviderInfo {
	return map[string]types.ProviderInfo{
		"godaddy": {
			Name:        "godaddy",
			DisplayName: "GoDaddy",
			Description: "World's largest domain registrar with comprehensive API support",
			DocumentationURL: "https://developer.godaddy.com/getstarted",
			Fields: []types.ProviderFieldInfo{
				{
					Name:        "api_key",
					DisplayName: "API Key",
					Type:        "text",
					Required:    true,
					Description: "Your GoDaddy API key",
					Placeholder: "3mM44UdWyeo_46cc991d7d9bcc9a_46cc991d7d9bcc9a",
				},
				{
					Name:        "api_secret",
					DisplayName: "API Secret",
					Type:        "password",
					Required:    true,
					Description: "Your GoDaddy API secret",
					Placeholder: "46cc991d7d9bcc9a46cc991d7d9bcc9a",
				},
			},
		},
		"namecheap": {
			Name:        "namecheap",
			DisplayName: "Namecheap",
			Description: "Popular domain registrar with competitive pricing and good API",
			DocumentationURL: "https://www.namecheap.com/support/api/intro/",
			Fields: []types.ProviderFieldInfo{
				{
					Name:        "api_key",
					DisplayName: "API Key",
					Type:        "text",
					Required:    true,
					Description: "Your Namecheap API key",
					Placeholder: "1234567890abcdef1234567890abcdef",
				},
				{
					Name:        "username",
					DisplayName: "API Username",
					Type:        "text",
					Required:    true,
					Description: "Your Namecheap API username (usually your account username)",
					Placeholder: "yourusername",
				},
				{
					Name:        "client_ip",
					DisplayName: "Client IP (Optional)",
					Type:        "text",
					Required:    false,
					Description: "Your server IP address (for API whitelisting)",
					Placeholder: "192.168.1.100",
				},
			},
		},
		"cloudflare": {
			Name:        "cloudflare",
			DisplayName: "Cloudflare Registrar",
			Description: "Cloudflare's registrar service with at-cost pricing",
			DocumentationURL: "https://developers.cloudflare.com/registrar/",
			Fields: []types.ProviderFieldInfo{
				{
					Name:        "api_token",
					DisplayName: "API Token",
					Type:        "password",
					Required:    true,
					Description: "Your Cloudflare API token with Registrar permissions",
					Placeholder: "1234567890abcdef1234567890abcdef12345678",
				},
				{
					Name:        "account_id",
					DisplayName: "Account ID",
					Type:        "text",
					Required:    true,
					Description: "Your Cloudflare account ID",
					Placeholder: "1234567890abcdef1234567890abcdef",
				},
			},
		},
		"mock": {
			Name:        "mock",
			DisplayName: "Mock Provider (Testing)",
			Description: "Mock provider for testing and development purposes",
			Fields: []types.ProviderFieldInfo{
				{
					Name:        "domain_count",
					DisplayName: "Number of Test Domains",
					Type:        "text",
					Required:    false,
					Description: "Number of mock domains to generate (default: 3)",
					Placeholder: "3",
				},
			},
		},
	}
}

// ValidateCredentials validates that all required credentials are provided
func (ps *ProviderService) ValidateCredentials(provider string, credentials map[string]string) error {
	providerInfo, exists := ps.supportedProviders[provider]
	if !exists {
		return fmt.Errorf("provider %s not supported", provider)
	}

	for _, field := range providerInfo.Fields {
		if field.Required {
			value, exists := credentials[field.Name]
			if !exists || value == "" {
				return fmt.Errorf("missing required field: %s (%s)", field.Name, field.DisplayName)
			}
		}
	}

	return nil
}

// GetProviderDisplayName returns the display name for a provider
func (ps *ProviderService) GetProviderDisplayName(provider string) string {
	if info, exists := ps.supportedProviders[provider]; exists {
		return info.DisplayName
	}
	return provider
}

// LogProviderConnection logs a connection attempt
func (ps *ProviderService) LogProviderConnection(provider, accountName string, success bool, message string) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}
	
	log.Printf("Provider Connection [%s] %s (%s): %s", 
		status, ps.GetProviderDisplayName(provider), accountName, message)
}
