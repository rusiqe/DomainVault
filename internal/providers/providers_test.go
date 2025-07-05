package providers

import (
	"testing"
	"time"

	"github.com/rusiqe/domainvault/internal/types"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		creds    ProviderCredentials
		wantErr  error
		wantType string
	}{
		{
			name:     "create mock client",
			provider: "mock",
			creds:    ProviderCredentials{},
			wantErr:  nil,
			wantType: "*providers.MockClient",
		},
		{
			name:     "create godaddy client",
			provider: "godaddy",
			creds: ProviderCredentials{
				"api_key":    "test-key",
				"api_secret": "test-secret",
			},
			wantErr:  nil,
			wantType: "*providers.GoDaddyClient",
		},
		{
			name:     "create namecheap client",
			provider: "namecheap",
			creds: ProviderCredentials{
				"api_key":  "test-key",
				"username": "test-user",
			},
			wantErr:  nil,
			wantType: "*providers.NamecheapClient",
		},
		{
			name:     "unsupported provider",
			provider: "unsupported",
			creds:    ProviderCredentials{},
			wantErr:  types.ErrUnsupportedProvider,
			wantType: "",
		},
		{
			name:     "godaddy missing api_key",
			provider: "godaddy",
			creds: ProviderCredentials{
				"api_secret": "test-secret",
			},
			wantErr:  types.ErrMissingConfig,
			wantType: "",
		},
		{
			name:     "namecheap missing username",
			provider: "namecheap",
			creds: ProviderCredentials{
				"api_key": "test-key",
			},
			wantErr:  types.ErrMissingConfig,
			wantType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.provider, tt.creds)
			
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewClient() unexpected error: %v", err)
				return
			}
			
			if client == nil {
				t.Errorf("NewClient() returned nil client")
				return
			}
			
			// Verify provider name
			if client.GetProviderName() != tt.provider {
				t.Errorf("GetProviderName() = %v, want %v", client.GetProviderName(), tt.provider)
			}
		})
	}
}

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		creds    ProviderCredentials
		wantErr  error
	}{
		{
			name:     "valid godaddy credentials",
			provider: "godaddy",
			creds: ProviderCredentials{
				"api_key":    "test-key",
				"api_secret": "test-secret",
			},
			wantErr: nil,
		},
		{
			name:     "godaddy missing api_key",
			provider: "godaddy",
			creds: ProviderCredentials{
				"api_secret": "test-secret",
			},
			wantErr: types.ErrMissingConfig,
		},
		{
			name:     "godaddy missing api_secret",
			provider: "godaddy",
			creds: ProviderCredentials{
				"api_key": "test-key",
			},
			wantErr: types.ErrMissingConfig,
		},
		{
			name:     "valid namecheap credentials",
			provider: "namecheap",
			creds: ProviderCredentials{
				"api_key":  "test-key",
				"username": "test-user",
			},
			wantErr: nil,
		},
		{
			name:     "namecheap missing api_key",
			provider: "namecheap",
			creds: ProviderCredentials{
				"username": "test-user",
			},
			wantErr: types.ErrMissingConfig,
		},
		{
			name:     "namecheap missing username",
			provider: "namecheap",
			creds: ProviderCredentials{
				"api_key": "test-key",
			},
			wantErr: types.ErrMissingConfig,
		},
		{
			name:     "mock credentials (no validation)",
			provider: "mock",
			creds:    ProviderCredentials{},
			wantErr:  nil,
		},
		{
			name:     "unsupported provider",
			provider: "unsupported",
			creds:    ProviderCredentials{},
			wantErr:  types.ErrUnsupportedProvider,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCredentials(tt.provider, tt.creds)
			if err != tt.wantErr {
				t.Errorf("ValidateCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMockClient(t *testing.T) {
	client, err := NewMockClient(ProviderCredentials{})
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	// Test GetProviderName
	if client.GetProviderName() != "mock" {
		t.Errorf("GetProviderName() = %v, want 'mock'", client.GetProviderName())
	}

	// Test FetchDomains
	domains, err := client.FetchDomains()
	if err != nil {
		t.Errorf("FetchDomains() unexpected error: %v", err)
	}

	if len(domains) == 0 {
		t.Error("FetchDomains() returned no domains")
	}

	// Verify mock domains have correct provider
	for _, domain := range domains {
		if domain.Provider != "mock" {
			t.Errorf("Domain provider = %v, want 'mock'", domain.Provider)
		}
		
		if domain.Name == "" {
			t.Error("Domain name is empty")
		}
		
		if domain.ID == "" {
			t.Error("Domain ID is empty")
		}
		
		if domain.ExpiresAt.IsZero() {
			t.Error("Domain expiration date is zero")
		}
	}

	// Test AddMockDomain
	newDomain := types.Domain{
		Name:      "test.com",
		ExpiresAt: time.Now().AddDate(1, 0, 0),
	}
	
	originalCount := len(domains)
	client.AddMockDomain(newDomain)
	
	updatedDomains, err := client.FetchDomains()
	if err != nil {
		t.Errorf("FetchDomains() after AddMockDomain unexpected error: %v", err)
	}
	
	if len(updatedDomains) != originalCount+1 {
		t.Errorf("Expected %d domains after adding, got %d", originalCount+1, len(updatedDomains))
	}
	
	// Verify the added domain has correct provider and ID
	found := false
	for _, domain := range updatedDomains {
		if domain.Name == "test.com" {
			found = true
			if domain.Provider != "mock" {
				t.Errorf("Added domain provider = %v, want 'mock'", domain.Provider)
			}
			if domain.ID == "" {
				t.Error("Added domain ID is empty")
			}
		}
	}
	
	if !found {
		t.Error("Added domain not found in updated list")
	}
}

func TestMockClient_Concurrency(t *testing.T) {
	// Test that mock client can handle concurrent requests
	client, err := NewMockClient(ProviderCredentials{})
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	// Run multiple goroutines calling FetchDomains
	results := make(chan []types.Domain, 5)
	errors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func() {
			domains, err := client.FetchDomains()
			if err != nil {
				errors <- err
				return
			}
			results <- domains
		}()
	}

	// Collect results
	for i := 0; i < 5; i++ {
		select {
		case domains := <-results:
			if len(domains) == 0 {
				t.Error("Concurrent FetchDomains() returned no domains")
			}
		case err := <-errors:
			t.Errorf("Concurrent FetchDomains() error: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent FetchDomains()")
		}
	}
}

func TestMockClient_SimulatedDelay(t *testing.T) {
	// Test that mock client simulates realistic API delays
	client, err := NewMockClient(ProviderCredentials{})
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	start := time.Now()
	_, err = client.FetchDomains()
	duration := time.Since(start)

	if err != nil {
		t.Errorf("FetchDomains() unexpected error: %v", err)
	}

	// Should take at least 100ms due to simulated delay
	if duration < 100*time.Millisecond {
		t.Errorf("FetchDomains() completed too quickly: %v", duration)
	}

	// But shouldn't take too long (safety check)
	if duration > 1*time.Second {
		t.Errorf("FetchDomains() took too long: %v", duration)
	}
}
