package providers

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rusiqe/domainvault/internal/storage"
	"github.com/rusiqe/domainvault/internal/types"
)

// SecureProviderService manages secure provider connections using environment variable references
// This service improves security by not storing API keys directly in the database
type SecureProviderService struct {
	repository             storage.DomainRepository
	connectedProviders     map[string]*SecureConnectedProvider
	autoSyncScheduler      *SecureAutoSyncScheduler
	mu                     sync.RWMutex
}

// SecureConnectedProvider represents a connected provider with secure credentials
type SecureConnectedProvider struct {
	ID                  string
	Provider            string
	Name                string
	AccountName         string
	CredentialReference string // Reference to environment variables
	Client              RegistrarClient
	Enabled             bool
	AutoSyncEnabled     bool
	SyncInterval        time.Duration
	LastSyncTime        time.Time
	LastSyncStatus      string
	ConnectionStatus    string
	DomainsCount        int
	ErrorCount          int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// SecureAutoSyncScheduler manages automatic syncing for secure providers
type SecureAutoSyncScheduler struct {
	providers  map[string]*SecureConnectedProvider
	syncFunc   func(providerID string) error
	tickers    map[string]*time.Ticker
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	running    bool
}

// NewSecureProviderService creates a new secure provider service
func NewSecureProviderService(repository storage.DomainRepository) *SecureProviderService {
	ctx, cancel := context.WithCancel(context.Background())
	
	service := &SecureProviderService{
		repository:         repository,
		connectedProviders: make(map[string]*SecureConnectedProvider),
		autoSyncScheduler: &SecureAutoSyncScheduler{
			providers: make(map[string]*SecureConnectedProvider),
			tickers:   make(map[string]*time.Ticker),
			ctx:       ctx,
			cancel:    cancel,
		},
	}
	
	// Load existing secure connections from database
	if err := service.loadExistingConnections(); err != nil {
		log.Printf("Warning: Failed to load existing secure connections: %v", err)
	}
	
	return service
}

// loadExistingConnections loads secure credential connections from the database
func (s *SecureProviderService) loadExistingConnections() error {
	credentials, err := s.repository.GetAllSecureCredentials()
	if err != nil {
		return fmt.Errorf("failed to load secure credentials: %w", err)
	}
	
	for _, cred := range credentials {
		if !cred.Enabled {
			continue
		}
		
		// Resolve credentials from environment variables
		resolvedCreds, err := types.ResolveCredentials(cred.CredentialReference)
		if err != nil {
			log.Printf("Warning: Failed to resolve credentials for %s (%s): %v", cred.Name, cred.CredentialReference, err)
			continue
		}
		
		// Convert to ProviderCredentials format for client creation
		providerCreds := make(ProviderCredentials)
		for key, value := range resolvedCreds {
			providerCreds[key] = value
		}
		
		// Create client
		client, err := NewClient(cred.Provider, providerCreds)
		if err != nil {
			log.Printf("Warning: Failed to create client for %s: %v", cred.Name, err)
			continue
		}
		
		// Create secure connected provider
		connectedProvider := &SecureConnectedProvider{
			ID:                  cred.ID,
			Provider:            cred.Provider,
			Name:                cred.Name,
			AccountName:         cred.AccountName,
			CredentialReference: cred.CredentialReference,
			Client:              client,
			Enabled:             cred.Enabled,
			AutoSyncEnabled:     true, // Default to enabled
			SyncInterval:        24 * time.Hour, // Default interval
			ConnectionStatus:    cred.ConnectionStatus,
			CreatedAt:           cred.CreatedAt,
			UpdatedAt:           cred.UpdatedAt,
		}
		
		s.connectedProviders[cred.ID] = connectedProvider
		
		// Add to auto-sync scheduler if enabled
		if connectedProvider.AutoSyncEnabled {
			s.autoSyncScheduler.AddProvider(connectedProvider)
		}
	}
	
	log.Printf("Loaded %d secure provider connections", len(s.connectedProviders))
	return nil
}

// GetAvailableCredentialOptions returns available credential options for a provider
func (s *SecureProviderService) GetAvailableCredentialOptions(provider string) []types.PredefinedCredentialOption {
	return types.GetCredentialOptions(provider)
}

// TestSecureConnection tests a connection using environment variable references
func (s *SecureProviderService) TestSecureConnection(provider, credentialReference string) (*types.ProviderConnectionResponse, error) {
	// Resolve credentials from environment variables
	credentials, err := types.ResolveCredentials(credentialReference)
	if err != nil {
		return &types.ProviderConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to resolve credentials: %v", err),
		}, nil
	}
	
	// Convert to ProviderCredentials format
	providerCreds := make(ProviderCredentials)
	for key, value := range credentials {
		providerCreds[key] = value
	}
	
	// Create test client
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

// AddSecureConnection adds a new secure provider connection
func (s *SecureProviderService) AddSecureConnection(req *types.SecureProviderConnectionRequest) (*types.ProviderConnectionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Test connection first if requested
	if req.TestConnection {
		testResult, err := s.TestSecureConnection(req.Provider, req.CredentialReference)
		if err != nil {
			return nil, fmt.Errorf("connection test failed: %w", err)
		}
		if !testResult.Success {
			return testResult, nil
		}
	}
	
	// Create secure credentials record
	secureCredentials := &types.SecureProviderCredentials{
		Provider:            req.Provider,
		Name:                req.Name,
		AccountName:         req.AccountName,
		CredentialReference: req.CredentialReference,
		Enabled:             true,
		ConnectionStatus:    "connected",
	}
	
	// Save to database
	if err := s.repository.CreateSecureCredentials(secureCredentials); err != nil {
		return &types.ProviderConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to save secure credentials: %v", err),
		}, nil
	}
	
	// Resolve credentials from environment variables for client creation
	resolvedCreds, err := types.ResolveCredentials(req.CredentialReference)
	if err != nil {
		return &types.ProviderConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to resolve credentials: %v", err),
		}, nil
	}
	
	// Convert to ProviderCredentials format
	providerCreds := make(ProviderCredentials)
	for key, value := range resolvedCreds {
		providerCreds[key] = value
	}
	
	// Create client
	client, err := NewClient(req.Provider, providerCreds)
	if err != nil {
		return &types.ProviderConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create client: %v", err),
		}, nil
	}
	
	// Create connected provider
	connectedProvider := &SecureConnectedProvider{
		ID:                  secureCredentials.ID,
		Provider:            req.Provider,
		Name:                req.Name,  
		AccountName:         req.AccountName,
		CredentialReference: req.CredentialReference,
		Client:              client,
		Enabled:             true,
		AutoSyncEnabled:     req.AutoSync,
		SyncInterval:        time.Duration(req.SyncIntervalHours) * time.Hour,
		ConnectionStatus:    "connected",
		CreatedAt:           secureCredentials.CreatedAt,
		UpdatedAt:           secureCredentials.UpdatedAt,
	}
	
	// Set default sync interval if not provided
	if connectedProvider.SyncInterval == 0 {
		connectedProvider.SyncInterval = 24 * time.Hour
	}
	
	// Add to connected providers
	s.connectedProviders[connectedProvider.ID] = connectedProvider
	
	// Add to auto-sync scheduler if enabled
	if connectedProvider.AutoSyncEnabled {
		s.autoSyncScheduler.AddProvider(connectedProvider)
	}
	
	response := &types.ProviderConnectionResponse{
		Success:    true,
		Message:    "Provider connected successfully using secure credentials",
		ProviderID: connectedProvider.ID,
	}
	
	// Run initial sync if requested
	if req.AutoSync {
		// TODO: Implement initial sync
		response.SyncStarted = true
	}
	
	log.Printf("Added secure provider connection: %s (%s) using reference %s", req.Name, req.Provider, req.CredentialReference)
	return response, nil
}

// GetSecureConnectedProviders returns all secure connected providers
func (s *SecureProviderService) GetSecureConnectedProviders() []*SecureConnectedProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	providers := make([]*SecureConnectedProvider, 0, len(s.connectedProviders))
	for _, provider := range s.connectedProviders {
		providers = append(providers, provider)
	}
	return providers
}

// GetSecureConnectedProvider returns a specific secure connected provider
func (s *SecureProviderService) GetSecureConnectedProvider(id string) (*SecureConnectedProvider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	provider, exists := s.connectedProviders[id]
	if !exists {
		return nil, fmt.Errorf("secure connected provider %s not found", id)
	}
	return provider, nil
}

// RemoveSecureConnection removes a secure provider connection
func (s *SecureProviderService) RemoveSecureConnection(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	provider, exists := s.connectedProviders[id]
	if !exists {
		return fmt.Errorf("secure connected provider %s not found", id)
	}
	
	// Remove from auto-sync scheduler
	s.autoSyncScheduler.RemoveProvider(id)
	
	// Remove from database
	if err := s.repository.DeleteSecureCredentials(id); err != nil {
		return fmt.Errorf("failed to delete secure credentials: %w", err)
	}
	
	// Remove from connected providers
	delete(s.connectedProviders, id)
	
	log.Printf("Removed secure provider connection: %s (%s)", provider.Name, provider.Provider)
	return nil
}

// SyncSecureProvider syncs a specific secure provider
func (s *SecureProviderService) SyncSecureProvider(id string, syncFunc func(RegistrarClient) ([]types.Domain, error)) error {
	s.mu.RLock()
	provider, exists := s.connectedProviders[id]
	s.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("secure connected provider %s not found", id)
	}
	
	if !provider.Enabled {
		return fmt.Errorf("provider %s is disabled", provider.Name)
	}
	
	log.Printf("Starting sync for secure provider: %s (%s)", provider.Name, provider.Provider)
	
	// Update sync status
	s.mu.Lock()
	provider.LastSyncTime = time.Now()
	provider.LastSyncStatus = "syncing"
	s.mu.Unlock()
	
	// Perform sync
	domains, err := syncFunc(provider.Client)
	if err != nil {
		s.mu.Lock()
		provider.LastSyncStatus = fmt.Sprintf("failed: %v", err)
		provider.ErrorCount++
		s.mu.Unlock()
		
		// Update database record
		s.updateProviderSyncStatus(id, "error", err.Error())
		return err
	}
	
	// Update sync status
	s.mu.Lock()
	provider.LastSyncStatus = "success"
	provider.DomainsCount = len(domains)
	provider.UpdatedAt = time.Now()
	s.mu.Unlock()
	
	// Update database record
	s.updateProviderSyncStatus(id, "connected", "")
	
	log.Printf("Sync completed for secure provider: %s (%s) - %d domains", provider.Name, provider.Provider, len(domains))
	return nil
}

// updateProviderSyncStatus updates the sync status in the database
func (s *SecureProviderService) updateProviderSyncStatus(id, status, errorMsg string) {
	creds, err := s.repository.GetSecureCredentialsByID(id)
	if err != nil {
		log.Printf("Warning: Failed to get secure credentials for status update: %v", err)
		return
	}
	
	creds.ConnectionStatus = status
	now := time.Now()
	creds.LastSync = &now
	if errorMsg != "" {
		creds.LastSyncError = &errorMsg
	} else {
		creds.LastSyncError = nil
	}
	
	if err := s.repository.UpdateSecureCredentials(creds); err != nil {
		log.Printf("Warning: Failed to update secure credentials status: %v", err)
	}
}

// StartAutoSync starts the auto-sync scheduler for secure providers
func (s *SecureProviderService) StartAutoSync(syncFunc func(RegistrarClient) ([]types.Domain, error)) {
	s.autoSyncScheduler.mu.Lock()
	defer s.autoSyncScheduler.mu.Unlock()
	
	if s.autoSyncScheduler.running {
		return
	}
	
	s.autoSyncScheduler.syncFunc = func(providerID string) error {
		return s.SyncSecureProvider(providerID, syncFunc)
	}
	
	s.autoSyncScheduler.running = true
	log.Println("Secure auto-sync scheduler started")
}

// StopAutoSync stops the auto-sync scheduler
func (s *SecureProviderService) StopAutoSync() {
	s.autoSyncScheduler.mu.Lock()
	defer s.autoSyncScheduler.mu.Unlock()
	
	if !s.autoSyncScheduler.running {
		return
	}
	
	// Stop all tickers
	for _, ticker := range s.autoSyncScheduler.tickers {
		ticker.Stop()
	}
	s.autoSyncScheduler.tickers = make(map[string]*time.Ticker)
	
	// Cancel context
	s.autoSyncScheduler.cancel()
	
	s.autoSyncScheduler.running = false
	log.Println("Secure auto-sync scheduler stopped")
}

// AddProvider adds a provider to the secure auto-sync scheduler
func (scheduler *SecureAutoSyncScheduler) AddProvider(provider *SecureConnectedProvider) {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	
	if !scheduler.running || !provider.AutoSyncEnabled {
		return
	}
	
	// Remove existing ticker if any
	if ticker, exists := scheduler.tickers[provider.ID]; exists {
		ticker.Stop()
	}
	
	// Create new ticker
	ticker := time.NewTicker(provider.SyncInterval)
	scheduler.tickers[provider.ID] = ticker
	scheduler.providers[provider.ID] = provider
	
	// Start goroutine for this provider
	go func(providerID string, ticker *time.Ticker) {
		for {
			select {
			case <-scheduler.ctx.Done():
				return
			case <-ticker.C:
				if scheduler.syncFunc != nil {
					if err := scheduler.syncFunc(providerID); err != nil {
						log.Printf("Secure auto-sync failed for provider %s: %v", providerID, err)
					} else {
						log.Printf("Secure auto-sync completed for provider %s", providerID)
					}
				}
			}
		}
	}(provider.ID, ticker)
	
	log.Printf("Added secure provider %s to auto-sync scheduler (interval: %v)", provider.Name, provider.SyncInterval)
}

// RemoveProvider removes a provider from the secure auto-sync scheduler
func (scheduler *SecureAutoSyncScheduler) RemoveProvider(providerID string) {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	
	if ticker, exists := scheduler.tickers[providerID]; exists {
		ticker.Stop()
		delete(scheduler.tickers, providerID)
	}
	
	delete(scheduler.providers, providerID)
	log.Printf("Removed secure provider %s from auto-sync scheduler", providerID)
}

// GetAutoSyncStatus returns the status of auto-sync for secure providers
func (s *SecureProviderService) GetAutoSyncStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	s.autoSyncScheduler.mu.RLock()
	defer s.autoSyncScheduler.mu.RUnlock()
	
	status := map[string]interface{}{
		"running":           s.autoSyncScheduler.running,
		"total_providers":   len(s.connectedProviders),
		"enabled_providers": 0,
		"syncing_providers": len(s.autoSyncScheduler.providers),
		"providers":         make([]map[string]interface{}, 0),
	}
	
	for _, provider := range s.connectedProviders {
		if provider.Enabled {
			status["enabled_providers"] = status["enabled_providers"].(int) + 1
		}
		
		providerStatus := map[string]interface{}{
			"id":                provider.ID,
			"name":              provider.Name,
			"provider":          provider.Provider,
			"enabled":           provider.Enabled,
			"auto_sync_enabled": provider.AutoSyncEnabled,
			"sync_interval":     provider.SyncInterval.String(),
			"last_sync_time":    provider.LastSyncTime,
			"last_sync_status":  provider.LastSyncStatus,
			"domains_count":     provider.DomainsCount,
			"error_count":       provider.ErrorCount,
			"credential_reference": provider.CredentialReference, // Show which env vars are being used
		}
		
		status["providers"] = append(status["providers"].([]map[string]interface{}), providerStatus)
	}
	
	return status
}

// Close cleans up resources
func (s *SecureProviderService) Close() error {
	s.StopAutoSync()
	return nil
}
