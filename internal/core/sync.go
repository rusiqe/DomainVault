package core

import (
	"fmt"
	"log"
	"sync"

	"github.com/rusiqe/domainvault/internal/providers"
	"github.com/rusiqe/domainvault/internal/storage"
	"github.com/rusiqe/domainvault/internal/types"
)

// SyncService manages domain synchronization across multiple providers
type SyncService struct {
	providers map[string]providers.RegistrarClient
	repo      storage.DomainRepository
	mu        sync.RWMutex // Protects providers map
}

// NewSyncService creates a new sync service
func NewSyncService(repo storage.DomainRepository) *SyncService {
	return &SyncService{
		providers: make(map[string]providers.RegistrarClient),
		repo:      repo,
	}
}

// AddProvider adds a registrar client to the sync service
func (s *SyncService) AddProvider(name string, client providers.RegistrarClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[name] = client
	log.Printf("Added provider: %s", name)
}

// RemoveProvider removes a registrar client from the sync service
func (s *SyncService) RemoveProvider(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.providers, name)
	log.Printf("Removed provider: %s", name)
}

// GetProviders returns a list of configured provider names
func (s *SyncService) GetProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	names := make([]string, 0, len(s.providers))
	for name := range s.providers {
		names = append(names, name)
	}
	return names
}

// Run executes a full synchronization across all providers
func (s *SyncService) Run() error {
	s.mu.RLock()
	providerCount := len(s.providers)
	s.mu.RUnlock()

	if providerCount == 0 {
		log.Println("No providers configured, skipping sync")
		return nil
	}

	log.Printf("Starting sync for %d providers", providerCount)

	// Channel to collect results from all providers
	results := make(chan SyncResult, providerCount)

	// Start goroutines for each provider
	s.mu.RLock()
	for name, client := range s.providers {
		go s.syncProvider(name, client, results)
	}
	s.mu.RUnlock()

	// Collect all domains from all providers
	var allDomains []types.Domain
	var errors []error

	for i := 0; i < providerCount; i++ {
		result := <-results
		if result.Error != nil {
			log.Printf("Provider %s sync failed: %v", result.ProviderName, result.Error)
			errors = append(errors, result.Error)
		} else {
			log.Printf("Provider %s fetched %d domains", result.ProviderName, len(result.Domains))
			allDomains = append(allDomains, result.Domains...)
		}
	}

	// Store all domains in the database
	if len(allDomains) > 0 {
		if err := s.repo.UpsertDomains(allDomains); err != nil {
			return fmt.Errorf("failed to store domains: %w", err)
		}
		log.Printf("Successfully synced %d domains total", len(allDomains))
	}

	// Return combined error if any providers failed
	if len(errors) > 0 {
		return fmt.Errorf("sync completed with %d provider errors", len(errors))
	}

	return nil
}

// SyncProvider synchronizes domains from a specific provider
func (s *SyncService) SyncProvider(providerName string) error {
	s.mu.RLock()
	client, exists := s.providers[providerName]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("provider %s not found", providerName)
	}

	log.Printf("Starting sync for provider: %s", providerName)

	domains, err := client.FetchDomains()
	if err != nil {
		return fmt.Errorf("failed to fetch domains from %s: %w", providerName, err)
	}

	if len(domains) == 0 {
		log.Printf("Provider %s returned no domains", providerName)
		return nil
	}

	// Store domains in database
	if err := s.repo.UpsertDomains(domains); err != nil {
		return fmt.Errorf("failed to store domains from %s: %w", providerName, err)
	}

	log.Printf("Successfully synced %d domains from %s", len(domains), providerName)
	return nil
}

// GetStatus returns the current sync service status
func (s *SyncService) GetStatus() SyncStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := SyncStatus{
		ProvidersConfigured: len(s.providers),
		Providers:          make(map[string]ProviderStatus),
	}

	for name := range s.providers {
		status.Providers[name] = ProviderStatus{
			Name:    name,
			Enabled: true,
			// LastSync and other fields would be populated from database/cache
		}
	}

	return status
}

// syncProvider is a helper function that runs in a goroutine
func (s *SyncService) syncProvider(name string, client providers.RegistrarClient, results chan<- SyncResult) {
	domains, err := client.FetchDomains()
	results <- SyncResult{
		ProviderName: name,
		Domains:      domains,
		Error:        err,
	}
}

// SyncResult represents the result of a provider sync operation
type SyncResult struct {
	ProviderName string
	Domains      []types.Domain
	Error        error
}

// SyncStatus represents the current status of the sync service
type SyncStatus struct {
	ProvidersConfigured int                       `json:"providers_configured"`
	Providers          map[string]ProviderStatus `json:"providers"`
}

// ProviderStatus represents the status of an individual provider
type ProviderStatus struct {
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	LastSync string `json:"last_sync,omitempty"`
	Error    string `json:"error,omitempty"`
}
