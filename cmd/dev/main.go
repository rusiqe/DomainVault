package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusiqe/domainvault/internal/analytics"
	"github.com/rusiqe/domainvault/internal/api"
	"github.com/rusiqe/domainvault/internal/auth"
	"github.com/rusiqe/domainvault/internal/core"
	"github.com/rusiqe/domainvault/internal/dns"
	"github.com/rusiqe/domainvault/internal/notifications"
	"github.com/rusiqe/domainvault/internal/providers"
	"github.com/rusiqe/domainvault/internal/security"
	"github.com/rusiqe/domainvault/internal/types"
)

// InMemoryRepo implements DomainRepository for development
type InMemoryRepo struct {
	domains []types.Domain
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		domains: make([]types.Domain, 0),
	}
}

func (r *InMemoryRepo) UpsertDomains(domains []types.Domain) error {
	for _, domain := range domains {
		// Check if domain exists
		found := false
		for i, existing := range r.domains {
			if existing.Name == domain.Name {
				r.domains[i] = domain
				found = true
				break
			}
		}
		if !found {
			r.domains = append(r.domains, domain)
		}
	}
	return nil
}

func (r *InMemoryRepo) GetAll() ([]types.Domain, error) {
	return r.domains, nil
}

func (r *InMemoryRepo) GetByID(id string) (*types.Domain, error) {
	for _, domain := range r.domains {
		if domain.ID == id {
			return &domain, nil
		}
	}
	return nil, types.ErrDomainNotFound
}

func (r *InMemoryRepo) GetByFilter(filter types.DomainFilter) ([]types.Domain, error) {
	result := make([]types.Domain, 0)
	
	for _, domain := range r.domains {
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

func (r *InMemoryRepo) Delete(id string) error {
	for i, domain := range r.domains {
		if domain.ID == id {
			r.domains = append(r.domains[:i], r.domains[i+1:]...)
			return nil
		}
	}
	return types.ErrDomainNotFound
}

func (r *InMemoryRepo) GetExpiring(threshold time.Duration) ([]types.Domain, error) {
	cutoff := time.Now().Add(threshold)
	result := make([]types.Domain, 0)
	
	for _, domain := range r.domains {
		if domain.ExpiresAt.Before(cutoff) {
			result = append(result, domain)
		}
	}
	
	return result, nil
}

func (r *InMemoryRepo) GetSummary() (*types.DomainSummary, error) {
	summary := &types.DomainSummary{
		Total:      len(r.domains),
		ByProvider: make(map[string]int),
		ExpiringIn: make(map[string]int),
		LastSync:   time.Now(),
	}
	
	for _, domain := range r.domains {
		summary.ByProvider[domain.Provider]++
	}
	
	// Count expiring domains
	now := time.Now()
	for _, domain := range r.domains {
		if domain.ExpiresAt.Before(now.AddDate(0, 0, 30)) {
			summary.ExpiringIn["30_days"]++
		}
		if domain.ExpiresAt.Before(now.AddDate(0, 0, 90)) {
			summary.ExpiringIn["90_days"]++
		}
		if domain.ExpiresAt.Before(now.AddDate(1, 0, 0)) {
			summary.ExpiringIn["365_days"]++
		}
	}
	
	return summary, nil
}

func (r *InMemoryRepo) BulkRenew(domainIDs []string) error {
	return nil // Mock implementation
}

func (r *InMemoryRepo) Close() error {
	return nil
}

func (r *InMemoryRepo) Ping() error {
	return nil
}

// Update method for domains
func (r *InMemoryRepo) Update(domain *types.Domain) error {
	for i, existing := range r.domains {
		if existing.ID == domain.ID {
			r.domains[i] = *domain
			return nil
		}
	}
	return types.ErrDomainNotFound
}

// GetDomainsByName returns domains matching the given domain name
func (r *InMemoryRepo) GetDomainsByName(name string) ([]types.Domain, error) {
	result := make([]types.Domain, 0)
	for _, domain := range r.domains {
		if domain.Name == name {
			result = append(result, domain)
		}
	}
	return result, nil
}

// User repository methods (in-memory)
func (r *InMemoryRepo) CreateUser(user *types.User) error {
	return nil // Mock implementation
}

func (r *InMemoryRepo) GetUserByUsername(username string) (*types.User, error) {
	if username == "admin" {
		return &types.User{
			ID:           "admin-id",
			Username:     "admin",
			Email:        "admin@domainvault.local",
			PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj0GdRLXu6G6", // admin123
			Role:         "admin",
			Enabled:      true,
		}, nil
	}
	return nil, types.ErrDomainNotFound
}

func (r *InMemoryRepo) GetUserByID(id string) (*types.User, error) {
	return nil, types.ErrDomainNotFound
}

func (r *InMemoryRepo) UpdateUser(user *types.User) error {
	return nil
}

func (r *InMemoryRepo) DeleteUser(id string) error {
	return nil
}

func (r *InMemoryRepo) CreateSession(session *types.Session) error {
	return nil
}

func (r *InMemoryRepo) GetSessionByToken(token string) (*types.Session, error) {
	return &types.Session{
		ID:        "session-id",
		UserID:    "admin-id",
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (r *InMemoryRepo) DeleteSession(token string) error {
	return nil
}

func (r *InMemoryRepo) DeleteExpiredSessions() error {
	return nil
}

func (r *InMemoryRepo) UpdateLastLogin(userID string) error {
	return nil
}

// DNS repository methods (in-memory)
func (r *InMemoryRepo) CreateRecord(record *types.DNSRecord) error {
	return nil
}

func (r *InMemoryRepo) GetRecordsByDomain(domainID string) ([]types.DNSRecord, error) {
	return []types.DNSRecord{
		{ID: "1", DomainID: domainID, Type: "A", Name: "@", Value: "192.168.1.1", TTL: 3600},
		{ID: "2", DomainID: domainID, Type: "A", Name: "www", Value: "192.168.1.1", TTL: 3600},
		{ID: "3", DomainID: domainID, Type: "MX", Name: "@", Value: "mail.example.com", TTL: 3600, Priority: &[]int{10}[0]},
	}, nil
}

func (r *InMemoryRepo) GetRecordByID(id string) (*types.DNSRecord, error) {
	return &types.DNSRecord{
		ID:       id,
		DomainID: "domain-1",
		Type:     "A",
		Name:     "@",
		Value:    "192.168.1.1",
		TTL:      3600,
	}, nil
}

func (r *InMemoryRepo) UpdateRecord(record *types.DNSRecord) error {
	return nil
}

func (r *InMemoryRepo) DeleteRecord(id string) error {
	return nil
}

func (r *InMemoryRepo) DeleteRecordsByDomain(domainID string) error {
	return nil
}

func (r *InMemoryRepo) BulkCreateRecords(records []types.DNSRecord) error {
	return nil
}

// Category and Project methods (in-memory)
func (r *InMemoryRepo) CreateCategory(category *types.Category) error {
	return nil
}

func (r *InMemoryRepo) GetAllCategories() ([]types.Category, error) {
	return []types.Category{
		{ID: "1", Name: "Personal", Description: "Personal domains", Color: "#6366f1"},
		{ID: "2", Name: "Business", Description: "Business domains", Color: "#dc2626"},
	}, nil
}

func (r *InMemoryRepo) GetCategoryByID(id string) (*types.Category, error) {
	return &types.Category{ID: id, Name: "Personal", Description: "Personal domains", Color: "#6366f1"}, nil
}

func (r *InMemoryRepo) UpdateCategory(category *types.Category) error {
	return nil
}

func (r *InMemoryRepo) DeleteCategory(id string) error {
	return nil
}

func (r *InMemoryRepo) CreateProject(project *types.Project) error {
	return nil
}

func (r *InMemoryRepo) GetAllProjects() ([]types.Project, error) {
	return []types.Project{
		{ID: "1", Name: "Portfolio Sites", Description: "Personal portfolio websites", Color: "#059669"},
		{ID: "2", Name: "E-commerce", Description: "Online store projects", Color: "#dc2626"},
	}, nil
}

func (r *InMemoryRepo) GetProjectByID(id string) (*types.Project, error) {
	return &types.Project{ID: id, Name: "Portfolio Sites", Description: "Personal portfolio websites", Color: "#059669"}, nil
}

func (r *InMemoryRepo) UpdateProject(project *types.Project) error {
	return nil
}

func (r *InMemoryRepo) DeleteProject(id string) error {
	return nil
}

// Credentials methods (in-memory)
func (r *InMemoryRepo) CreateCredentials(creds *types.ProviderCredentials) error {
	return nil
}

func (r *InMemoryRepo) GetAllCredentials() ([]types.ProviderCredentials, error) {
	return []types.ProviderCredentials{
		{ID: "1", Provider: "mock", Name: "Mock Provider", Enabled: true},
	}, nil
}

func (r *InMemoryRepo) GetCredentialsByID(id string) (*types.ProviderCredentials, error) {
	return &types.ProviderCredentials{ID: id, Provider: "mock", Name: "Mock Provider", Enabled: true}, nil
}

func (r *InMemoryRepo) GetCredentialsByProvider(provider string) ([]types.ProviderCredentials, error) {
	return []types.ProviderCredentials{
		{ID: "1", Provider: provider, Name: "Mock Provider", Enabled: true},
	}, nil
}

func (r *InMemoryRepo) UpdateCredentials(creds *types.ProviderCredentials) error {
	return nil
}

func (r *InMemoryRepo) DeleteCredentials(id string) error {
	return nil
}

func main() {
	fmt.Println("🚀 DomainVault Development Server")
	fmt.Println("=====================================")
	fmt.Println("📦 Using in-memory storage")
	fmt.Println("🔧 Mock provider enabled")
	fmt.Println("🌐 Server will start on http://localhost:8080")
	fmt.Println("")

	// Initialize in-memory storage
	repo := NewInMemoryRepo()

	// Initialize services
	syncSvc := core.NewSyncService(repo)
	authSvc := auth.NewAuthService(repo)
	dnsSvc := dns.NewDNSService(repo)
	analyticsSvc := analytics.NewAnalyticsService(repo)
	
	// Mock notification configurations
	emailConfig := notifications.EmailConfig{Enabled: false}
	slackConfig := notifications.SlackConfig{Enabled: false}
	webhookConfig := notifications.WebhookConfig{Enabled: false}
	notificationSvc := notifications.NewNotificationService(emailConfig, slackConfig, webhookConfig)
	
	// Mock security service - use nil for development
	var securitySvc *security.SecurityService = nil

	// Add mock provider
	mockClient, err := providers.NewMockClient(providers.ProviderCredentials{})
	if err != nil {
		log.Fatal("Failed to create mock client:", err)
	}
	syncSvc.AddProvider("mock", mockClient)

	// Perform initial sync to populate data
	fmt.Println("🔄 Performing initial sync...")
	if err := syncSvc.Run(); err != nil {
		log.Printf("Initial sync error: %v", err)
	} else {
		fmt.Println("✅ Initial sync completed")
	}

	// Start periodic sync
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := syncSvc.Run(); err != nil {
				log.Printf("Periodic sync failed: %v", err)
			} else {
				log.Println("🔄 Periodic sync completed")
			}
		}
	}()

	// Initialize API handlers
	handler := api.NewDomainHandler(repo, syncSvc)
	adminHandler := api.NewAdminHandler(repo, authSvc, syncSvc, dnsSvc, analyticsSvc, notificationSvc, securitySvc)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Add CORS middleware for frontend testing
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	handler.RegisterRoutes(r)
	adminHandler.RegisterAdminRoutes(r)

	// Serve static files
	r.Static("/static", "./web/static")

	// Serve the main frontend
	r.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	// Serve admin interface
	r.GET("/admin", func(c *gin.Context) {
		c.File("./web/admin.html")
	})

	// Add a simple API documentation endpoint
	r.GET("/dev", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, `
<!DOCTYPE html>
<html>
<head>
    <title>DomainVault - Development</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #2c3e50; border-bottom: 3px solid #3498db; padding-bottom: 10px; }
        .status { background: #e8f5e8; padding: 10px; border-radius: 4px; margin: 10px 0; }
        .endpoint { background: #f8f9fa; padding: 10px; margin: 10px 0; border-left: 4px solid #007bff; }
        code { background: #f1f2f6; padding: 2px 6px; border-radius: 3px; font-family: 'Courier New', monospace; }
        .button { display: inline-block; padding: 8px 16px; margin: 5px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; }
        .button:hover { background: #0056b3; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🏰 DomainVault Development Server</h1>
        
        <div class="status">
			✅ Server is running on <strong>http://localhost:8080</strong><br>
			🗄️ Using in-memory storage (data will reset on restart)<br>
			🔧 Mock provider active (generates sample domain data)<br>
			🔐 Admin interface available at <a href="/admin" target="_blank">/admin</a> (admin/admin123)
        </div>

        <h2>🌐 API Endpoints</h2>
        
        <div class="endpoint">
            <strong>GET /api/v1/health</strong> - Health check<br>
            <a href="/api/v1/health" class="button" target="_blank">Test Health</a>
        </div>

        <div class="endpoint">
            <strong>GET /api/v1/domains</strong> - List all domains<br>
            <a href="/api/v1/domains" class="button" target="_blank">View Domains</a>
        </div>

        <div class="endpoint">
            <strong>GET /api/v1/domains/summary</strong> - Domain statistics<br>
            <a href="/api/v1/domains/summary" class="button" target="_blank">View Summary</a>
        </div>

        <div class="endpoint">
            <strong>GET /api/v1/domains/expiring</strong> - Expiring domains<br>
            <a href="/api/v1/domains/expiring?days=30" class="button" target="_blank">Expiring in 30 days</a>
        </div>

        <div class="endpoint">
            <strong>POST /api/v1/sync</strong> - Trigger manual sync<br>
            <button onclick="triggerSync()" class="button">Trigger Sync</button>
        </div>

        <div class="endpoint">
            <strong>GET /api/v1/sync/status</strong> - Sync service status<br>
            <a href="/api/v1/sync/status" class="button" target="_blank">View Status</a>
        </div>

        <h2>🧪 Manual Testing</h2>
        <p>Use the buttons above to test the API endpoints, or use curl commands:</p>
        <code>curl http://localhost:8080/api/v1/domains</code><br>
        <code>curl -X POST http://localhost:8080/api/v1/sync</code><br>
        <code>curl http://localhost:8080/api/v1/domains?provider=mock</code>

        <div id="result" style="margin-top: 20px; padding: 10px; background: #f8f9fa; border-radius: 4px; display: none;">
            <h3>Response:</h3>
            <pre id="response"></pre>
        </div>
    </div>

    <script>
        async function triggerSync() {
            try {
                const response = await fetch('/api/v1/sync', { method: 'POST' });
                const data = await response.json();
                document.getElementById('response').textContent = JSON.stringify(data, null, 2);
                document.getElementById('result').style.display = 'block';
            } catch (error) {
                document.getElementById('response').textContent = 'Error: ' + error.message;
                document.getElementById('result').style.display = 'block';
            }
        }
    </script>
</body>
</html>`)
	})

	// Start server
	fmt.Printf("🌐 Starting server on http://localhost:8080\n")
	fmt.Printf("📝 Visit http://localhost:8080 for the development interface\n")
	fmt.Printf("🔐 Admin panel available at http://localhost:8080/admin\n")
	fmt.Printf("🔑 Admin credentials: admin / admin123\n")
	fmt.Printf("🔗 API available at http://localhost:8080/api/v1/\n\n")

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
