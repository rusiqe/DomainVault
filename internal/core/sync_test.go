package core

import (
	"errors"
	"testing"
	"time"

	"github.com/rusiqe/domainvault/internal/providers"
	"github.com/rusiqe/domainvault/internal/types"
)

// Mock repository for testing
type mockRepository struct {
	domains []types.Domain
	fail    bool
}

func (m *mockRepository) UpsertDomains(domains []types.Domain) error {
	if m.fail {
		return errors.New("database error")
	}
	m.domains = append(m.domains, domains...)
	return nil
}

func (m *mockRepository) GetAll() ([]types.Domain, error) {
	return m.domains, nil
}

func (m *mockRepository) GetByID(id string) (*types.Domain, error) {
	for _, domain := range m.domains {
		if domain.ID == id {
			return &domain, nil
		}
	}
	return nil, types.ErrDomainNotFound
}

func (m *mockRepository) GetByFilter(filter types.DomainFilter) ([]types.Domain, error) {
	return m.domains, nil
}

func (m *mockRepository) Delete(id string) error {
	return nil
}

func (m *mockRepository) GetExpiring(threshold time.Duration) ([]types.Domain, error) {
	return nil, nil
}

func (m *mockRepository) GetSummary() (*types.DomainSummary, error) {
	return &types.DomainSummary{}, nil
}

func (m *mockRepository) BulkRenew(domainIDs []string) error {
	return nil
}

func (m *mockRepository) Close() error {
	return nil
}

func (m *mockRepository) Ping() error {
	return nil
}

// Mock provider client for testing
type mockProviderClient struct {
	name    string
	domains []types.Domain
	fail    bool
}

func (m *mockProviderClient) FetchDomains() ([]types.Domain, error) {
	if m.fail {
		return nil, errors.New("provider error")
	}
	return m.domains, nil
}

func (m *mockProviderClient) GetProviderName() string {
	return m.name
}

func TestNewSyncService(t *testing.T) {
	repo := &mockRepository{}
	service := NewSyncService(repo)

	if service == nil {
		t.Fatal("NewSyncService() returned nil")
	}

	if service.repo != repo {
		t.Error("Repository not set correctly")
	}

	if len(service.providers) != 0 {
		t.Error("Expected empty providers map")
	}
}

func TestSyncService_AddProvider(t *testing.T) {
	repo := &mockRepository{}
	service := NewSyncService(repo)

	mockClient := &mockProviderClient{
		name: "test-provider",
		domains: []types.Domain{
			{Name: "test.com", Provider: "test-provider"},
		},
	}

	service.AddProvider("test-provider", mockClient)

	providers := service.GetProviders()
	if len(providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(providers))
	}

	if providers[0] != "test-provider" {
		t.Errorf("Expected 'test-provider', got %s", providers[0])
	}
}

func TestSyncService_RemoveProvider(t *testing.T) {
	repo := &mockRepository{}
	service := NewSyncService(repo)

	mockClient := &mockProviderClient{name: "test-provider"}
	service.AddProvider("test-provider", mockClient)

	// Verify provider is added
	if len(service.GetProviders()) != 1 {
		t.Error("Provider not added")
	}

	service.RemoveProvider("test-provider")

	// Verify provider is removed
	if len(service.GetProviders()) != 0 {
		t.Error("Provider not removed")
	}
}

func TestSyncService_Run_NoProviders(t *testing.T) {
	repo := &mockRepository{}
	service := NewSyncService(repo)

	err := service.Run()
	if err != nil {
		t.Errorf("Run() with no providers should not return error, got: %v", err)
	}

	if len(repo.domains) != 0 {
		t.Error("No domains should be stored when no providers configured")
	}
}

func TestSyncService_Run_SingleProvider(t *testing.T) {
	repo := &mockRepository{}
	service := NewSyncService(repo)

	mockClient := &mockProviderClient{
		name: "test-provider",
		domains: []types.Domain{
			{Name: "test1.com", Provider: "test-provider"},
			{Name: "test2.com", Provider: "test-provider"},
		},
	}

	service.AddProvider("test-provider", mockClient)

	err := service.Run()
	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
	}

	if len(repo.domains) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(repo.domains))
	}
}

func TestSyncService_Run_MultipleProviders(t *testing.T) {
	repo := &mockRepository{}
	service := NewSyncService(repo)

	// Add first provider
	mockClient1 := &mockProviderClient{
		name: "provider1",
		domains: []types.Domain{
			{Name: "test1.com", Provider: "provider1"},
		},
	}
	service.AddProvider("provider1", mockClient1)

	// Add second provider
	mockClient2 := &mockProviderClient{
		name: "provider2",
		domains: []types.Domain{
			{Name: "test2.com", Provider: "provider2"},
			{Name: "test3.com", Provider: "provider2"},
		},
	}
	service.AddProvider("provider2", mockClient2)

	err := service.Run()
	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
	}

	if len(repo.domains) != 3 {
		t.Errorf("Expected 3 domains, got %d", len(repo.domains))
	}
}

func TestSyncService_Run_ProviderError(t *testing.T) {
	repo := &mockRepository{}
	service := NewSyncService(repo)

	// Add working provider
	mockClient1 := &mockProviderClient{
		name: "good-provider",
		domains: []types.Domain{
			{Name: "test1.com", Provider: "good-provider"},
		},
	}
	service.AddProvider("good-provider", mockClient1)

	// Add failing provider
	mockClient2 := &mockProviderClient{
		name: "bad-provider",
		fail: true,
	}
	service.AddProvider("bad-provider", mockClient2)

	err := service.Run()
	if err == nil {
		t.Error("Run() should return error when provider fails")
	}

	// Should still store domains from working provider
	if len(repo.domains) != 1 {
		t.Errorf("Expected 1 domain from working provider, got %d", len(repo.domains))
	}
}

func TestSyncService_Run_DatabaseError(t *testing.T) {
	repo := &mockRepository{fail: true}
	service := NewSyncService(repo)

	mockClient := &mockProviderClient{
		name: "test-provider",
		domains: []types.Domain{
			{Name: "test.com", Provider: "test-provider"},
		},
	}
	service.AddProvider("test-provider", mockClient)

	err := service.Run()
	if err == nil {
		t.Error("Run() should return error when database fails")
	}
}

func TestSyncService_SyncProvider(t *testing.T) {
	repo := &mockRepository{}
	service := NewSyncService(repo)

	mockClient := &mockProviderClient{
		name: "test-provider",
		domains: []types.Domain{
			{Name: "test.com", Provider: "test-provider"},
		},
	}
	service.AddProvider("test-provider", mockClient)

	err := service.SyncProvider("test-provider")
	if err != nil {
		t.Errorf("SyncProvider() unexpected error: %v", err)
	}

	if len(repo.domains) != 1 {
		t.Errorf("Expected 1 domain, got %d", len(repo.domains))
	}
}

func TestSyncService_SyncProvider_NotFound(t *testing.T) {
	repo := &mockRepository{}
	service := NewSyncService(repo)

	err := service.SyncProvider("nonexistent-provider")
	if err == nil {
		t.Error("SyncProvider() should return error for nonexistent provider")
	}
}

func TestSyncService_SyncProvider_NoDomains(t *testing.T) {
	repo := &mockRepository{}
	service := NewSyncService(repo)

	mockClient := &mockProviderClient{
		name:    "empty-provider",
		domains: []types.Domain{}, // No domains
	}
	service.AddProvider("empty-provider", mockClient)

	err := service.SyncProvider("empty-provider")
	if err != nil {
		t.Errorf("SyncProvider() with no domains should not error: %v", err)
	}

	if len(repo.domains) != 0 {
		t.Errorf("Expected 0 domains, got %d", len(repo.domains))
	}
}

func TestSyncService_GetStatus(t *testing.T) {
	repo := &mockRepository{}
	service := NewSyncService(repo)

	// Initially no providers
	status := service.GetStatus()
	if status.ProvidersConfigured != 0 {
		t.Errorf("Expected 0 providers configured, got %d", status.ProvidersConfigured)
	}

	// Add providers
	mockClient1 := &mockProviderClient{name: "provider1"}
	mockClient2 := &mockProviderClient{name: "provider2"}
	
	service.AddProvider("provider1", mockClient1)
	service.AddProvider("provider2", mockClient2)

	status = service.GetStatus()
	if status.ProvidersConfigured != 2 {
		t.Errorf("Expected 2 providers configured, got %d", status.ProvidersConfigured)
	}

	if len(status.Providers) != 2 {
		t.Errorf("Expected 2 provider statuses, got %d", len(status.Providers))
	}

	// Check individual provider status
	if status.Providers["provider1"].Name != "provider1" {
		t.Error("Provider1 status not correct")
	}

	if !status.Providers["provider1"].Enabled {
		t.Error("Provider1 should be enabled")
	}
}

func TestSyncService_Concurrency(t *testing.T) {
	repo := &mockRepository{}
	service := NewSyncService(repo)

	// Add mock provider
	mockClient := &mockProviderClient{
		name: "test-provider",
		domains: []types.Domain{
			{Name: "test.com", Provider: "test-provider"},
		},
	}
	service.AddProvider("test-provider", mockClient)

	// Run multiple sync operations concurrently
	errChan := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func() {
			errChan <- service.SyncProvider("test-provider")
		}()
	}

	// Collect results
	for i := 0; i < 5; i++ {
		select {
		case err := <-errChan:
			if err != nil {
				t.Errorf("Concurrent sync error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent sync")
		}
	}
}

func TestSyncService_Integration_WithMockProvider(t *testing.T) {
	// This tests the integration with the actual mock provider
	repo := &mockRepository{}
	service := NewSyncService(repo)

	// Use the real mock provider
	mockClient, err := providers.NewMockClient(providers.ProviderCredentials{})
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	service.AddProvider("mock", mockClient)

	err = service.Run()
	if err != nil {
		t.Errorf("Integration sync failed: %v", err)
	}

	// Mock provider should return some domains
	if len(repo.domains) == 0 {
		t.Error("Expected domains from mock provider")
	}

	// Verify all domains have correct provider
	for _, domain := range repo.domains {
		if domain.Provider != "mock" {
			t.Errorf("Expected provider 'mock', got %s", domain.Provider)
		}
	}
}
