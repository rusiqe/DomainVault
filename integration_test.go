package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusiqe/domainvault/internal/api"
	"github.com/rusiqe/domainvault/internal/core"
	"github.com/rusiqe/domainvault/internal/providers"
	"github.com/rusiqe/domainvault/internal/types"
)

// Mock repository for integration testing
type mockIntegrationRepo struct {
	domains []types.Domain
}

func (m *mockIntegrationRepo) UpsertDomains(domains []types.Domain) error {
	m.domains = append(m.domains, domains...)
	return nil
}

func (m *mockIntegrationRepo) GetAll() ([]types.Domain, error) {
	return m.domains, nil
}

func (m *mockIntegrationRepo) GetByID(id string) (*types.Domain, error) {
	for _, domain := range m.domains {
		if domain.ID == id {
			return &domain, nil
		}
	}
	return nil, types.ErrDomainNotFound
}

func (m *mockIntegrationRepo) GetByFilter(filter types.DomainFilter) ([]types.Domain, error) {
	result := make([]types.Domain, 0)
	
	for _, domain := range m.domains {
		// Apply filters
		if filter.Provider != "" && domain.Provider != filter.Provider {
			continue
		}
		
		if filter.Search != "" && domain.Name != filter.Search {
			continue
		}
		
		if filter.ExpiresAfter != nil && domain.ExpiresAt.Before(*filter.ExpiresAfter) {
			continue
		}
		
		if filter.ExpiresBefore != nil && domain.ExpiresAt.After(*filter.ExpiresBefore) {
			continue
		}
		
		result = append(result, domain)
	}
	
	// Apply pagination
	start := filter.Offset
	if start > len(result) {
		return []types.Domain{}, nil
	}
	
	end := start + filter.Limit
	if filter.Limit == 0 || end > len(result) {
		end = len(result)
	}
	
	return result[start:end], nil
}

func (m *mockIntegrationRepo) Delete(id string) error {
	for i, domain := range m.domains {
		if domain.ID == id {
			m.domains = append(m.domains[:i], m.domains[i+1:]...)
			return nil
		}
	}
	return types.ErrDomainNotFound
}

func (m *mockIntegrationRepo) GetExpiring(threshold time.Duration) ([]types.Domain, error) {
	cutoff := time.Now().Add(threshold)
	result := make([]types.Domain, 0)
	
	for _, domain := range m.domains {
		if domain.ExpiresAt.Before(cutoff) {
			result = append(result, domain)
		}
	}
	
	return result, nil
}

func (m *mockIntegrationRepo) GetSummary() (*types.DomainSummary, error) {
	summary := &types.DomainSummary{
		Total:      len(m.domains),
		ByProvider: make(map[string]int),
		ExpiringIn: make(map[string]int),
		LastSync:   time.Now(),
	}
	
	for _, domain := range m.domains {
		summary.ByProvider[domain.Provider]++
	}
	
	return summary, nil
}

func (m *mockIntegrationRepo) BulkRenew(domainIDs []string) error {
	return nil
}

func (m *mockIntegrationRepo) Close() error {
	return nil
}

func (m *mockIntegrationRepo) Ping() error {
	return nil
}

func TestIntegration_FullWorkflow(t *testing.T) {
	// Setup components
	repo := &mockIntegrationRepo{}
	syncSvc := core.NewSyncService(repo)
	
	// Add mock provider
	mockClient, err := providers.NewMockClient(providers.ProviderCredentials{})
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}
	syncSvc.AddProvider("mock", mockClient)
	
	// Setup API handler
	handler := api.NewDomainHandler(repo, syncSvc)
	
	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.RegisterRoutes(router)
	
	// Test 1: Health check
	t.Run("Health Check", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		
		if resp.Code != http.StatusOK {
			t.Errorf("Health check failed. Status: %d", resp.Code)
		}
		
		var response map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &response)
		
		if response["status"] != "healthy" {
			t.Error("Expected healthy status")
		}
	})
	
	// Test 2: Initial empty state
	t.Run("Initial Empty State", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/domains", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		
		if resp.Code != http.StatusOK {
			t.Errorf("List domains failed. Status: %d", resp.Code)
		}
		
		var response map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &response)
		
		domains := response["domains"].([]interface{})
		if len(domains) != 0 {
			t.Errorf("Expected 0 domains initially, got %d", len(domains))
		}
	})
	
	// Test 3: Trigger sync
	t.Run("Trigger Sync", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/sync", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		
		if resp.Code != http.StatusAccepted {
			t.Errorf("Trigger sync failed. Status: %d", resp.Code)
		}
		
		// Wait a bit for async sync to complete
		time.Sleep(200 * time.Millisecond)
	})
	
	// Test 4: Check domains after sync
	t.Run("Domains After Sync", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/domains", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		
		if resp.Code != http.StatusOK {
			t.Errorf("List domains failed. Status: %d", resp.Code)
		}
		
		var response map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &response)
		
		domains := response["domains"].([]interface{})
		if len(domains) == 0 {
			t.Error("Expected domains after sync, got 0")
		}
		
		// Verify domain structure
		firstDomain := domains[0].(map[string]interface{})
		if firstDomain["provider"] != "mock" {
			t.Error("Expected provider to be 'mock'")
		}
		
		if firstDomain["name"] == "" {
			t.Error("Domain name should not be empty")
		}
	})
	
	// Test 5: Get summary
	t.Run("Get Summary", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/domains/summary", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		
		if resp.Code != http.StatusOK {
			t.Errorf("Get summary failed. Status: %d", resp.Code)
		}
		
		var summary types.DomainSummary
		json.Unmarshal(resp.Body.Bytes(), &summary)
		
		if summary.Total == 0 {
			t.Error("Expected non-zero total domains in summary")
		}
		
		if summary.ByProvider["mock"] == 0 {
			t.Error("Expected mock provider to have domains")
		}
	})
	
	// Test 6: Filter by provider
	t.Run("Filter By Provider", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/domains?provider=mock", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		
		if resp.Code != http.StatusOK {
			t.Errorf("Filter by provider failed. Status: %d", resp.Code)
		}
		
		var response map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &response)
		
		domains := response["domains"].([]interface{})
		for _, domain := range domains {
			d := domain.(map[string]interface{})
			if d["provider"] != "mock" {
				t.Error("All filtered domains should have provider 'mock'")
			}
		}
	})
	
	// Test 7: Get sync status
	t.Run("Get Sync Status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/sync/status", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		
		if resp.Code != http.StatusOK {
			t.Errorf("Get sync status failed. Status: %d", resp.Code)
		}
		
		var status core.SyncStatus
		json.Unmarshal(resp.Body.Bytes(), &status)
		
		if status.ProvidersConfigured == 0 {
			t.Error("Expected configured providers")
		}
		
		if status.Providers["mock"].Name != "mock" {
			t.Error("Expected mock provider in status")
		}
	})
	
	// Test 8: Get expiring domains
	t.Run("Get Expiring Domains", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/domains/expiring?days=365", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		
		if resp.Code != http.StatusOK {
			t.Errorf("Get expiring domains failed. Status: %d", resp.Code)
		}
		
		var response map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &response)
		
		// Should have at least some domains expiring within 365 days
		count := response["count"].(float64)
		if count == 0 {
			t.Error("Expected some domains to be expiring within 365 days")
		}
	})
}

func TestIntegration_ErrorHandling(t *testing.T) {
	// Setup components
	repo := &mockIntegrationRepo{}
	syncSvc := core.NewSyncService(repo)
	handler := api.NewDomainHandler(repo, syncSvc)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.RegisterRoutes(router)
	
	// Test 1: Get non-existent domain
	t.Run("Get Non-existent Domain", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/domains/non-existent-id", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		
		if resp.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", resp.Code)
		}
	})
	
	// Test 2: Delete non-existent domain
	t.Run("Delete Non-existent Domain", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v1/domains/non-existent-id", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		
		if resp.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", resp.Code)
		}
	})
	
	// Test 3: Sync non-existent provider
	t.Run("Sync Non-existent Provider", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/sync/non-existent", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		
		// Should still accept the request but fail in background
		if resp.Code != http.StatusAccepted {
			t.Errorf("Expected 202, got %d", resp.Code)
		}
	})
}

func TestIntegration_ConcurrentRequests(t *testing.T) {
	// Setup components
	repo := &mockIntegrationRepo{}
	syncSvc := core.NewSyncService(repo)
	
	// Add mock provider
	mockClient, err := providers.NewMockClient(providers.ProviderCredentials{})
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}
	syncSvc.AddProvider("mock", mockClient)
	
	handler := api.NewDomainHandler(repo, syncSvc)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.RegisterRoutes(router)
	
	// Trigger initial sync
	syncSvc.Run()
	
	// Test concurrent GET requests
	t.Run("Concurrent GET Requests", func(t *testing.T) {
		results := make(chan int, 10)
		
		for i := 0; i < 10; i++ {
			go func() {
				req, _ := http.NewRequest("GET", "/api/v1/domains", nil)
				resp := httptest.NewRecorder()
				router.ServeHTTP(resp, req)
				results <- resp.Code
			}()
		}
		
		// Collect results
		for i := 0; i < 10; i++ {
			select {
			case code := <-results:
				if code != http.StatusOK {
					t.Errorf("Concurrent request failed with status %d", code)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent requests")
			}
		}
	})
}
