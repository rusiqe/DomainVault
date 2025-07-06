package providers

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rusiqe/domainvault/internal/types"
)

// ProviderService manages provider information and connections
type ProviderService struct {
	supportedProviders map[string]types.ProviderInfo
	connectedProviders map[string]*ConnectedProvider
	autoSyncScheduler  *AutoSyncScheduler
	mu                 sync.RWMutex
}

// ConnectedProvider represents a connected provider with its credentials
type ConnectedProvider struct {
	ID               string
	Provider         string
	Name             string
	AccountName      string
	Credentials      ProviderCredentials
	Client           RegistrarClient
	Enabled          bool
	AutoSyncEnabled  bool
	SyncInterval     time.Duration
	LastSyncTime     time.Time
	LastSyncStatus   string
	ConnectionStatus string
	DomainsCount     int
	ErrorCount       int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// AutoSyncScheduler manages automatic syncing for providers
type AutoSyncScheduler struct {
	providers  map[string]*ConnectedProvider
	syncFunc   func(providerID string) error
	tickers    map[string]*time.Ticker
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	running    bool
}

// NewProviderService creates a new provider service
func NewProviderService() *ProviderService {
	ctx, cancel := context.WithCancel(context.Background())
	
	ps := &ProviderService{
		supportedProviders: initializeSupportedProviders(),
		connectedProviders: make(map[string]*ConnectedProvider),
		autoSyncScheduler: &AutoSyncScheduler{
			providers: make(map[string]*ConnectedProvider),
			tickers:   make(map[string]*time.Ticker),
			ctx:       ctx,
			cancel:    cancel,
		},
	}
	
	return ps
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

// ============================================================================
// ENHANCED PROVIDER MANAGEMENT
// ============================================================================

// AddConnectedProvider adds a new connected provider
func (ps *ProviderService) AddConnectedProvider(req *types.ProviderConnectionRequest) (*ConnectedProvider, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	// Validate provider is supported
	if !ps.IsSupported(req.Provider) {
		return nil, fmt.Errorf("provider %s not supported", req.Provider)
	}
	
	// Validate credentials
	if err := ps.ValidateCredentials(req.Provider, req.Credentials); err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}
	
	// Convert credentials
	providerCreds := make(ProviderCredentials)
	for key, value := range req.Credentials {
		providerCreds[key] = value
	}
	
	// Create client
	client, err := NewClient(req.Provider, providerCreds)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	
	// Test connection if requested
	if req.TestConnection {
		if _, err := client.FetchDomains(); err != nil {
			return nil, fmt.Errorf("connection test failed: %w", err)
		}
	}
	
	// Create connected provider
	connectedProvider := &ConnectedProvider{
		ID:               generateID(),
		Provider:         req.Provider,
		Name:             req.Name,
		AccountName:      req.AccountName,
		Credentials:      providerCreds,
		Client:           client,
		Enabled:          true,
		AutoSyncEnabled:  req.AutoSync,
		SyncInterval:     time.Duration(req.SyncIntervalHours) * time.Hour,
		ConnectionStatus: "connected",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	
	// Set default sync interval if not provided
	if connectedProvider.SyncInterval == 0 {
		connectedProvider.SyncInterval = 24 * time.Hour // Default to 24 hours
	}
	
	// Add to connected providers
	ps.connectedProviders[connectedProvider.ID] = connectedProvider
	
	// Add to auto-sync scheduler if enabled
	if connectedProvider.AutoSyncEnabled {
		ps.autoSyncScheduler.AddProvider(connectedProvider)
	}
	
	return connectedProvider, nil
}

// GetConnectedProviders returns all connected providers
func (ps *ProviderService) GetConnectedProviders() []*ConnectedProvider {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	providers := make([]*ConnectedProvider, 0, len(ps.connectedProviders))
	for _, provider := range ps.connectedProviders {
		providers = append(providers, provider)
	}
	return providers
}

// GetConnectedProvider returns a specific connected provider
func (ps *ProviderService) GetConnectedProvider(id string) (*ConnectedProvider, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	provider, exists := ps.connectedProviders[id]
	if !exists {
		return nil, fmt.Errorf("connected provider %s not found", id)
	}
	return provider, nil
}

// UpdateConnectedProvider updates a connected provider
func (ps *ProviderService) UpdateConnectedProvider(id string, updates map[string]interface{}) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	provider, exists := ps.connectedProviders[id]
	if !exists {
		return fmt.Errorf("connected provider %s not found", id)
	}
	
	// Update fields
	if name, ok := updates["name"].(string); ok {
		provider.Name = name
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		provider.Enabled = enabled
	}
	if autoSync, ok := updates["auto_sync_enabled"].(bool); ok {
		provider.AutoSyncEnabled = autoSync
		if autoSync {
			ps.autoSyncScheduler.AddProvider(provider)
		} else {
			ps.autoSyncScheduler.RemoveProvider(id)
		}
	}
	if interval, ok := updates["sync_interval_hours"].(float64); ok {
		provider.SyncInterval = time.Duration(interval) * time.Hour
		if provider.AutoSyncEnabled {
			ps.autoSyncScheduler.UpdateProvider(provider)
		}
	}
	
	provider.UpdatedAt = time.Now()
	return nil
}

// RemoveConnectedProvider removes a connected provider
func (ps *ProviderService) RemoveConnectedProvider(id string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	provider, exists := ps.connectedProviders[id]
	if !exists {
		return fmt.Errorf("connected provider %s not found", id)
	}
	
	// Remove from auto-sync scheduler
	ps.autoSyncScheduler.RemoveProvider(id)
	
	// Remove from connected providers
	delete(ps.connectedProviders, id)
	
	log.Printf("Removed connected provider: %s (%s)", provider.Name, provider.Provider)
	return nil
}

// SyncProvider syncs a specific provider
func (ps *ProviderService) SyncProvider(id string, syncFunc func(RegistrarClient) ([]types.Domain, error)) error {
	ps.mu.RLock()
	provider, exists := ps.connectedProviders[id]
	ps.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("connected provider %s not found", id)
	}
	
	if !provider.Enabled {
		return fmt.Errorf("provider %s is disabled", provider.Name)
	}
	
	log.Printf("Starting sync for provider: %s (%s)", provider.Name, provider.Provider)
	
	// Update sync status
	ps.mu.Lock()
	provider.LastSyncTime = time.Now()
	provider.LastSyncStatus = "syncing"
	ps.mu.Unlock()
	
	// Perform sync
	domains, err := syncFunc(provider.Client)
	if err != nil {
		ps.mu.Lock()
		provider.LastSyncStatus = fmt.Sprintf("failed: %v", err)
		provider.ErrorCount++
		ps.mu.Unlock()
		return err
	}
	
	// Update sync status
	ps.mu.Lock()
	provider.LastSyncStatus = "success"
	provider.DomainsCount = len(domains)
	provider.UpdatedAt = time.Now()
	ps.mu.Unlock()
	
	log.Printf("Sync completed for provider: %s (%s) - %d domains", provider.Name, provider.Provider, len(domains))
	return nil
}

// SyncAllProviders syncs all enabled providers
func (ps *ProviderService) SyncAllProviders(syncFunc func(RegistrarClient) ([]types.Domain, error)) error {
	ps.mu.RLock()
	providers := make([]*ConnectedProvider, 0, len(ps.connectedProviders))
	for _, provider := range ps.connectedProviders {
		if provider.Enabled {
			providers = append(providers, provider)
		}
	}
	ps.mu.RUnlock()
	
	log.Printf("Starting sync for %d enabled providers", len(providers))
	
	errors := make([]error, 0)
	for _, provider := range providers {
		if err := ps.SyncProvider(provider.ID, syncFunc); err != nil {
			errors = append(errors, fmt.Errorf("provider %s: %w", provider.Name, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("sync completed with %d errors: %v", len(errors), errors)
	}
	
	log.Printf("Sync completed successfully for all %d providers", len(providers))
	return nil
}

// ============================================================================
// AUTO-SYNC SCHEDULER
// ============================================================================

// StartAutoSync starts the auto-sync scheduler
func (ps *ProviderService) StartAutoSync(syncFunc func(RegistrarClient) ([]types.Domain, error)) {
	ps.autoSyncScheduler.mu.Lock()
	defer ps.autoSyncScheduler.mu.Unlock()
	
	if ps.autoSyncScheduler.running {
		return
	}
	
	ps.autoSyncScheduler.syncFunc = func(providerID string) error {
		return ps.SyncProvider(providerID, syncFunc)
	}
	
	ps.autoSyncScheduler.running = true
	log.Println("Auto-sync scheduler started")
}

// StopAutoSync stops the auto-sync scheduler
func (ps *ProviderService) StopAutoSync() {
	ps.autoSyncScheduler.mu.Lock()
	defer ps.autoSyncScheduler.mu.Unlock()
	
	if !ps.autoSyncScheduler.running {
		return
	}
	
	// Stop all tickers
	for _, ticker := range ps.autoSyncScheduler.tickers {
		ticker.Stop()
	}
	ps.autoSyncScheduler.tickers = make(map[string]*time.Ticker)
	
	// Cancel context
	ps.autoSyncScheduler.cancel()
	
	ps.autoSyncScheduler.running = false
	log.Println("Auto-sync scheduler stopped")
}

// AddProvider adds a provider to the auto-sync scheduler
func (scheduler *AutoSyncScheduler) AddProvider(provider *ConnectedProvider) {
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
						log.Printf("Auto-sync failed for provider %s: %v", providerID, err)
					} else {
						log.Printf("Auto-sync completed for provider %s", providerID)
					}
				}
			}
		}
	}(provider.ID, ticker)
	
	log.Printf("Added provider %s to auto-sync scheduler (interval: %v)", provider.Name, provider.SyncInterval)
}

// RemoveProvider removes a provider from the auto-sync scheduler
func (scheduler *AutoSyncScheduler) RemoveProvider(providerID string) {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	
	if ticker, exists := scheduler.tickers[providerID]; exists {
		ticker.Stop()
		delete(scheduler.tickers, providerID)
	}
	
	delete(scheduler.providers, providerID)
	log.Printf("Removed provider %s from auto-sync scheduler", providerID)
}

// UpdateProvider updates a provider in the auto-sync scheduler
func (scheduler *AutoSyncScheduler) UpdateProvider(provider *ConnectedProvider) {
	scheduler.RemoveProvider(provider.ID)
	scheduler.AddProvider(provider)
}

// GetAutoSyncStatus returns the status of auto-sync for all providers
func (ps *ProviderService) GetAutoSyncStatus() map[string]interface{} {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	ps.autoSyncScheduler.mu.RLock()
	defer ps.autoSyncScheduler.mu.RUnlock()
	
	status := map[string]interface{}{
		"running":           ps.autoSyncScheduler.running,
		"total_providers":   len(ps.connectedProviders),
		"enabled_providers": 0,
		"syncing_providers": len(ps.autoSyncScheduler.providers),
		"providers":         make([]map[string]interface{}, 0),
	}
	
	for _, provider := range ps.connectedProviders {
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
		}
		
		status["providers"] = append(status["providers"].([]map[string]interface{}), providerStatus)
	}
	
	return status
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// generateID generates a unique ID for providers
func generateID() string {
	return fmt.Sprintf("provider_%d", time.Now().UnixNano())
}
