package providers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/rusiqe/domainvault/internal/types"
)

// NamecheapClient implements RegistrarClient for Namecheap API
type NamecheapClient struct {
	apiKey   string
	username string
	baseURL  string
	client   *http.Client
}

// NamecheapResponse represents the XML response structure
type NamecheapResponse struct {
	Status      string               `xml:"Status,attr"`
	CommandResponse NamecheapCommandResponse `xml:"CommandResponse"`
	Errors      []NamecheapError     `xml:"Errors>Error"`
}

type NamecheapCommandResponse struct {
	DomainGetListResult NamecheapDomainList `xml:"DomainGetListResult"`
}

type NamecheapDomainList struct {
	Domains []NamecheapDomain `xml:"Domain"`
}

type NamecheapDomain struct {
	ID       int64  `xml:"ID,attr"`
	Name     string `xml:"Name,attr"`
	User     string `xml:"User,attr"`
	Created  string `xml:"Created,attr"`
	Expires  string `xml:"Expires,attr"`
	IsExpired bool  `xml:"IsExpired,attr"`
}

type NamecheapError struct {
	Number      string `xml:"Number,attr"`
	Description string `xml:",chardata"`
}

// NewNamecheapClient creates a new Namecheap client
func NewNamecheapClient(creds ProviderCredentials) (*NamecheapClient, error) {
	apiKey, ok := creds["api_key"].(string)
	if !ok {
		return nil, types.ErrMissingConfig
	}
	
	username, ok := creds["username"].(string)
	if !ok {
		return nil, types.ErrMissingConfig
	}

	return &NamecheapClient{
		apiKey:   apiKey,
		username: username,
		baseURL:  "https://api.namecheap.com/xml.response",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// FetchDomains retrieves domains from Namecheap API
func (n *NamecheapClient) FetchDomains() ([]types.Domain, error) {
	params := url.Values{}
	params.Set("ApiUser", n.username)
	params.Set("ApiKey", n.apiKey)
	params.Set("UserName", n.username)
	params.Set("Command", "namecheap.domains.getList")
	params.Set("ClientIp", "127.0.0.1") // Namecheap requires client IP
	
	url := fmt.Sprintf("%s?%s", n.baseURL, params.Encode())
	
	resp, err := n.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch domains: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, types.ErrProviderAuth
	}
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var ncResponse NamecheapResponse
	if err := xml.NewDecoder(resp.Body).Decode(&ncResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if ncResponse.Status != "OK" {
		if len(ncResponse.Errors) > 0 {
			return nil, fmt.Errorf("namecheap API error: %s", ncResponse.Errors[0].Description)
		}
		return nil, fmt.Errorf("unknown namecheap API error")
	}

	// Convert to internal domain format
	domains := make([]types.Domain, len(ncResponse.CommandResponse.DomainGetListResult.Domains))
	for i, nd := range ncResponse.CommandResponse.DomainGetListResult.Domains {
		// Parse dates
		createdAt, _ := time.Parse("01/02/2006", nd.Created)
		expiresAt, _ := time.Parse("01/02/2006", nd.Expires)
		
		domains[i] = types.Domain{
			ID:        uuid.New().String(), // Generate new UUID
			Name:      nd.Name,
			Provider:  "namecheap",
			ExpiresAt: expiresAt,
			CreatedAt: createdAt,
			UpdatedAt: time.Now(),
		}
	}

	return domains, nil
}

// GetProviderName returns the provider name
func (n *NamecheapClient) GetProviderName() string {
	return "namecheap"
}

// Future implementations for MVP expansion:
// func (n *NamecheapClient) RenewDomain(domainID string) error { ... }
// func (n *NamecheapClient) UpdateDNS(domain string, records []types.DNSRecord) error { ... }
