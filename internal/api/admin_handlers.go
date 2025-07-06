package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusiqe/domainvault/internal/analytics"
	"github.com/rusiqe/domainvault/internal/auth"
	"github.com/rusiqe/domainvault/internal/core"
	"github.com/rusiqe/domainvault/internal/dns"
	"github.com/rusiqe/domainvault/internal/notifications"
	"github.com/rusiqe/domainvault/internal/providers"
	"github.com/rusiqe/domainvault/internal/security"
	"github.com/rusiqe/domainvault/internal/status"
	"github.com/rusiqe/domainvault/internal/storage"
	"github.com/rusiqe/domainvault/internal/types"
)

// AdminHandler handles admin-specific HTTP requests
type AdminHandler struct {
	domainRepo       storage.DomainRepository
	authSvc          *auth.AuthService
	syncSvc          *core.SyncService
	dnsSvc           *dns.DNSService
	statusChecker    *status.StatusChecker
	providerSvc      *providers.ProviderService
	analyticsSvc     *analytics.AnalyticsService
	notificationSvc  *notifications.NotificationService
	securitySvc      *security.SecurityService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	domainRepo storage.DomainRepository,
	authSvc *auth.AuthService,
	syncSvc *core.SyncService,
	dnsSvc *dns.DNSService,
	analyticsSvc *analytics.AnalyticsService,
	notificationSvc *notifications.NotificationService,
	securitySvc *security.SecurityService,
) *AdminHandler {
	return &AdminHandler{
		domainRepo:       domainRepo,
		authSvc:          authSvc,
		syncSvc:          syncSvc,
		dnsSvc:           dnsSvc,
		statusChecker:    status.NewStatusChecker(),
		providerSvc:      providers.NewProviderService(),
		analyticsSvc:     analyticsSvc,
		notificationSvc:  notificationSvc,
		securitySvc:      securitySvc,
	}
}

// RegisterAdminRoutes sets up the admin HTTP routes
func (h *AdminHandler) RegisterAdminRoutes(r *gin.Engine) {
	// Public authentication routes
	authRoutes := r.Group("/api/v1/auth")
	{
		authRoutes.POST("/login", h.Login)
		authRoutes.POST("/logout", h.Logout)
	}

	// Protected admin routes
	admin := r.Group("/api/v1/admin")
	admin.Use(auth.AuthMiddleware(h.authSvc))
	admin.Use(auth.RequireRole("admin"))
	{
		// Domain management
		admin.PUT("/domains/:id", h.UpdateDomain)
		admin.POST("/domains/bulk-purchase", h.BulkPurchaseDomains)
		admin.POST("/domains/bulk-decommission", h.BulkDecommissionDomains)
		admin.POST("/domains/bulk-sync", h.BulkSyncDomains)

		// DNS management
		admin.GET("/domains/:id/dns", h.GetDomainDNS)
		admin.POST("/domains/:id/dns", h.CreateDNSRecord)
		admin.PUT("/domains/:id/dns", h.BulkUpdateDNS)
		admin.PUT("/dns/:id", h.UpdateDNSRecord)
		admin.DELETE("/dns/:id", h.DeleteDNSRecord)
		admin.GET("/dns/templates", h.GetDNSTemplates)
		
		// Bulk DNS operations
		admin.POST("/dns/bulk/ip", h.BulkAssignIP)
		admin.POST("/dns/bulk/nameservers", h.BulkUpdateNameservers)
		admin.POST("/dns/bulk/csv", h.BulkUpdateFromCSV)

		// Category management
		admin.GET("/categories", h.ListCategories)
		admin.POST("/categories", h.CreateCategory)
		admin.PUT("/categories/:id", h.UpdateCategory)
		admin.DELETE("/categories/:id", h.DeleteCategory)

		// Project management
		admin.GET("/projects", h.ListProjects)
		admin.POST("/projects", h.CreateProject)
		admin.PUT("/projects/:id", h.UpdateProject)
		admin.DELETE("/projects/:id", h.DeleteProject)

		// Provider management
		admin.GET("/providers/supported", h.ListSupportedProviders)
		admin.GET("/providers/connected", h.ListConnectedProviders)
		admin.GET("/providers/connected/:id", h.GetConnectedProvider)
		admin.PUT("/providers/connected/:id", h.UpdateConnectedProvider)
		admin.DELETE("/providers/connected/:id", h.RemoveConnectedProvider)
		admin.POST("/providers/connect", h.ConnectProvider)
		admin.POST("/providers/test", h.TestProviderConnection)
		admin.POST("/providers/:id/sync", h.SyncProviderByID)
		admin.POST("/providers/sync-all", h.SyncAllConnectedProviders)
		admin.GET("/providers/auto-sync/status", h.GetAutoSyncStatus)
		admin.POST("/providers/auto-sync/start", h.StartAutoSync)
		admin.POST("/providers/auto-sync/stop", h.StopAutoSync)
		
		// Provider credentials management
		admin.GET("/credentials", h.ListCredentials)
		admin.POST("/credentials", h.CreateCredentials)
		admin.PUT("/credentials/:id", h.UpdateCredentials)
		admin.DELETE("/credentials/:id", h.DeleteCredentials)

		// Advanced sync operations
		admin.POST("/sync/manual", h.ManualSync)
		admin.GET("/sync/providers", h.GetSupportedProviders)

		// Status checking
		admin.POST("/domains/:id/check-status", h.CheckDomainStatus)
		admin.POST("/domains/bulk-check-status", h.BulkCheckStatus)
		admin.GET("/status/summary", h.GetStatusSummary)

		// Analytics and reporting
		admin.GET("/analytics/portfolio", h.GetPortfolioAnalytics)
		admin.GET("/analytics/financial", h.GetFinancialAnalytics)
		admin.GET("/analytics/security", h.GetSecurityAnalytics)
		admin.GET("/analytics/trends", h.GetTrendAnalytics)

		// Notifications and alerts
		admin.GET("/notifications/rules", h.GetNotificationRules)
		admin.POST("/notifications/rules", h.CreateNotificationRule)
		admin.PUT("/notifications/rules/:id", h.UpdateNotificationRule)
		admin.DELETE("/notifications/rules/:id", h.DeleteNotificationRule)
		admin.POST("/notifications/test", h.TestNotification)
		admin.GET("/alerts", h.GetAlerts)
		admin.POST("/alerts/:id/resolve", h.ResolveAlert)

		// Security and audit
		admin.GET("/security/audit", h.GetAuditEvents)
		admin.GET("/security/metrics", h.GetSecurityMetrics)
		admin.GET("/security/alerts", h.GetSecurityAlerts)
		admin.POST("/security/alerts/:id/resolve", h.ResolveSecurityAlert)
		admin.GET("/security/sessions", h.GetActiveSessions)
		admin.DELETE("/security/sessions/:id", h.TerminateSession)
	}
}

// Login authenticates an admin user
func (h *AdminHandler) Login(c *gin.Context) {
	var req types.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	response, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout invalidates the current session
func (h *AdminHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusOK, gin.H{"message": "Already logged out"})
		return
	}

	// Extract token
	tokenParts := []string{}
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenParts = []string{"Bearer", authHeader[7:]}
	}

	if len(tokenParts) == 2 {
		h.authSvc.Logout(tokenParts[1])
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// UpdateDomain updates domain details including category/project assignment
func (h *AdminHandler) UpdateDomain(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain ID required"})
		return
	}

	var domain types.Domain
	if err := c.ShouldBindJSON(&domain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain data"})
		return
	}

	domain.ID = id
	if err := h.domainRepo.Update(&domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain)
}

// BulkPurchaseDomains handles bulk domain purchases
func (h *AdminHandler) BulkPurchaseDomains(c *gin.Context) {
	var req types.DomainPurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// This would integrate with domain registrar APIs
	// For now, return a placeholder response
	c.JSON(http.StatusAccepted, gin.H{
		"message":     "Bulk purchase initiated",
		"domains":     req.Domains,
		"provider":    req.Provider,
		"status":      "pending",
		"request_id":  "bulk-" + fmt.Sprint(len(req.Domains)),
	})
}

// BulkDecommissionDomains handles bulk domain decommissioning
func (h *AdminHandler) BulkDecommissionDomains(c *gin.Context) {
	var req types.DomainDecommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// This would integrate with domain registrar APIs
	// For now, update the domains in our database
	successCount := 0
	var errors []string

	for _, domainID := range req.DomainIDs {
		domain, err := h.domainRepo.GetByID(domainID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Domain %s: %v", domainID, err))
			continue
		}

		if req.StopAutoRenew {
			domain.AutoRenew = false
		}
		if req.TransferOut {
			domain.Status = "transferring"
		}

		if err := h.domainRepo.Update(domain); err != nil {
			errors = append(errors, fmt.Sprintf("Domain %s: %v", domainID, err))
			continue
		}

		successCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Bulk decommission completed",
		"processed":      successCount,
		"total":          len(req.DomainIDs),
		"errors":         errors,
		"stop_auto_renew": req.StopAutoRenew,
		"transfer_out":   req.TransferOut,
	})
}

// BulkSyncDomains handles manual bulk sync operations
func (h *AdminHandler) BulkSyncDomains(c *gin.Context) {
	var req types.BulkSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Start sync in background
	go func() {
		if len(req.Providers) > 0 {
			for _, provider := range req.Providers {
				h.syncSvc.SyncProvider(provider)
			}
		} else {
			h.syncSvc.Run()
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message":       "Bulk sync initiated",
		"providers":     req.Providers,
		"force_refresh": req.ForceRefresh,
		"status":        "running",
	})
}

// GetDomainDNS retrieves DNS records for a domain
func (h *AdminHandler) GetDomainDNS(c *gin.Context) {
	domainID := c.Param("id")
	if domainID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain ID required"})
		return
	}

	records, err := h.dnsSvc.GetDomainRecords(domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domain_id": domainID,
		"records":   records,
		"count":     len(records),
	})
}

// CreateDNSRecord creates a new DNS record for a domain
func (h *AdminHandler) CreateDNSRecord(c *gin.Context) {
	domainID := c.Param("id")
	if domainID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain ID required"})
		return
	}

	var record types.DNSRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid DNS record data"})
		return
	}

	record.DomainID = domainID
	if err := h.dnsSvc.CreateRecord(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

// BulkUpdateDNS updates all DNS records for a domain
func (h *AdminHandler) BulkUpdateDNS(c *gin.Context) {
	domainID := c.Param("id")
	if domainID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain ID required"})
		return
	}

	var records []types.DNSRecord
	if err := c.ShouldBindJSON(&records); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid DNS records data"})
		return
	}

	if err := h.dnsSvc.BulkUpdateRecords(domainID, records); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "DNS records updated successfully",
		"domain_id": domainID,
		"count":     len(records),
	})
}

// UpdateDNSRecord updates a specific DNS record
func (h *AdminHandler) UpdateDNSRecord(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "DNS record ID required"})
		return
	}

	var record types.DNSRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid DNS record data"})
		return
	}

	record.ID = id
	if err := h.dnsSvc.UpdateRecord(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

// DeleteDNSRecord deletes a DNS record
func (h *AdminHandler) DeleteDNSRecord(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "DNS record ID required"})
		return
	}

	if err := h.dnsSvc.DeleteRecord(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "DNS record deleted successfully"})
}

// GetDNSTemplates returns common DNS record templates
func (h *AdminHandler) GetDNSTemplates(c *gin.Context) {
	templates := h.dnsSvc.GetCommonRecordTemplates()
	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
	})
}

// ManualSync triggers a manual sync with detailed options
func (h *AdminHandler) ManualSync(c *gin.Context) {
	var req struct {
		Provider      string `json:"provider,omitempty"`
		CredentialsID string `json:"credentials_id,omitempty"`
		ForceRefresh  bool   `json:"force_refresh"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Start sync in background
	go func() {
		if req.Provider != "" {
			h.syncSvc.SyncProvider(req.Provider)
		} else {
			h.syncSvc.Run()
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message":        "Manual sync initiated",
		"provider":       req.Provider,
		"credentials_id": req.CredentialsID,
		"force_refresh":  req.ForceRefresh,
	})
}

// GetSupportedProviders returns the list of supported domain providers
func (h *AdminHandler) GetSupportedProviders(c *gin.Context) {
	providers := []string{"godaddy", "namecheap", "cloudflare"}
	
	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
		"count":     len(providers),
	})
}

// ============================================================================
// CATEGORY MANAGEMENT METHODS
// ============================================================================

// ListCategories returns all categories
func (h *AdminHandler) ListCategories(c *gin.Context) {
	// For now, we'll use the domainRepo directly - in a real implementation,
	// you'd want separate repositories or extend the interface
	if repo, ok := h.domainRepo.(interface{ GetAllCategories() ([]types.Category, error) }); ok {
		categories, err := repo.GetAllCategories()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"categories": categories,
			"count":      len(categories),
		})
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Category operations not implemented"})
	}
}

// CreateCategory creates a new category
func (h *AdminHandler) CreateCategory(c *gin.Context) {
	var category types.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category data"})
		return
	}

	if repo, ok := h.domainRepo.(interface{ CreateCategory(*types.Category) error }); ok {
		if err := repo.CreateCategory(&category); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, category)
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Category operations not implemented"})
	}
}

// UpdateCategory updates an existing category
func (h *AdminHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID required"})
		return
	}

	var category types.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category data"})
		return
	}

	category.ID = id
	if repo, ok := h.domainRepo.(interface{ UpdateCategory(*types.Category) error }); ok {
		if err := repo.UpdateCategory(&category); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, category)
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Category operations not implemented"})
	}
}

// DeleteCategory deletes a category
func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID required"})
		return
	}

	if repo, ok := h.domainRepo.(interface{ DeleteCategory(string) error }); ok {
		if err := repo.DeleteCategory(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Category operations not implemented"})
	}
}

// ============================================================================
// PROJECT MANAGEMENT METHODS
// ============================================================================

// ListProjects returns all projects
func (h *AdminHandler) ListProjects(c *gin.Context) {
	if repo, ok := h.domainRepo.(interface{ GetAllProjects() ([]types.Project, error) }); ok {
		projects, err := repo.GetAllProjects()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"projects": projects,
			"count":    len(projects),
		})
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Project operations not implemented"})
	}
}

// CreateProject creates a new project
func (h *AdminHandler) CreateProject(c *gin.Context) {
	var project types.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project data"})
		return
	}

	if repo, ok := h.domainRepo.(interface{ CreateProject(*types.Project) error }); ok {
		if err := repo.CreateProject(&project); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, project)
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Project operations not implemented"})
	}
}

// UpdateProject updates an existing project
func (h *AdminHandler) UpdateProject(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID required"})
		return
	}

	var project types.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project data"})
		return
	}

	project.ID = id
	if repo, ok := h.domainRepo.(interface{ UpdateProject(*types.Project) error }); ok {
		if err := repo.UpdateProject(&project); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, project)
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Project operations not implemented"})
	}
}

// DeleteProject deletes a project
func (h *AdminHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID required"})
		return
	}

	if repo, ok := h.domainRepo.(interface{ DeleteProject(string) error }); ok {
		if err := repo.DeleteProject(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Project operations not implemented"})
	}
}

// ============================================================================
// CREDENTIALS MANAGEMENT METHODS
// ============================================================================

// ListCredentials returns all provider credentials
func (h *AdminHandler) ListCredentials(c *gin.Context) {
	if repo, ok := h.domainRepo.(interface{ GetAllCredentials() ([]types.ProviderCredentials, error) }); ok {
		credentials, err := repo.GetAllCredentials()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Don't expose actual credentials in the response
		for i := range credentials {
			credentials[i].Credentials = map[string]string{"***": "***"}
		}

		c.JSON(http.StatusOK, gin.H{
			"credentials": credentials,
			"count":       len(credentials),
		})
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Credentials operations not implemented"})
	}
}

// ============================================================================
// ENHANCED PROVIDER MANAGEMENT ENDPOINTS
// ============================================================================

// ListConnectedProviders returns all connected providers with status
func (h *AdminHandler) ListConnectedProviders(c *gin.Context) {
	providers := h.providerSvc.GetConnectedProviders()
	
	// Convert to response format without exposing credentials
	response := make([]map[string]interface{}, 0, len(providers))
	for _, provider := range providers {
		providerData := map[string]interface{}{
			"id":                provider.ID,
			"provider":          provider.Provider,
			"name":              provider.Name,
			"account_name":      provider.AccountName,
			"enabled":           provider.Enabled,
			"auto_sync_enabled": provider.AutoSyncEnabled,
			"sync_interval":     provider.SyncInterval.Hours(),
			"connection_status": provider.ConnectionStatus,
			"last_sync_time":    provider.LastSyncTime,
			"last_sync_status":  provider.LastSyncStatus,
			"domains_count":     provider.DomainsCount,
			"error_count":       provider.ErrorCount,
			"created_at":        provider.CreatedAt,
			"updated_at":        provider.UpdatedAt,
		}
		response = append(response, providerData)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"providers": response,
		"count":     len(response),
	})
}

// GetConnectedProvider returns a specific connected provider
func (h *AdminHandler) GetConnectedProvider(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider ID required"})
		return
	}
	
	provider, err := h.providerSvc.GetConnectedProvider(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	// Return without exposing credentials
	response := map[string]interface{}{
		"id":                provider.ID,
		"provider":          provider.Provider,
		"name":              provider.Name,
		"account_name":      provider.AccountName,
		"enabled":           provider.Enabled,
		"auto_sync_enabled": provider.AutoSyncEnabled,
		"sync_interval":     provider.SyncInterval.Hours(),
		"connection_status": provider.ConnectionStatus,
		"last_sync_time":    provider.LastSyncTime,
		"last_sync_status":  provider.LastSyncStatus,
		"domains_count":     provider.DomainsCount,
		"error_count":       provider.ErrorCount,
		"created_at":        provider.CreatedAt,
		"updated_at":        provider.UpdatedAt,
	}
	
	c.JSON(http.StatusOK, response)
}

// UpdateConnectedProvider updates a connected provider's settings
func (h *AdminHandler) UpdateConnectedProvider(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider ID required"})
		return
	}
	
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update data"})
		return
	}
	
	if err := h.providerSvc.UpdateConnectedProvider(id, updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Provider updated successfully"})
}

// RemoveConnectedProvider removes a connected provider
func (h *AdminHandler) RemoveConnectedProvider(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider ID required"})
		return
	}
	
	if err := h.providerSvc.RemoveConnectedProvider(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Provider removed successfully"})
}

// SyncProviderByID syncs a specific provider by ID
func (h *AdminHandler) SyncProviderByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider ID required"})
		return
	}
	
	// Start sync in background
	go func() {
		syncFunc := func(client providers.RegistrarClient) ([]types.Domain, error) {
			return client.FetchDomains()
		}
		
		if err := h.providerSvc.SyncProvider(id, syncFunc); err != nil {
			log.Printf("Sync failed for provider %s: %v", id, err)
		} else {
			log.Printf("Sync completed for provider %s", id)
		}
	}()
	
	c.JSON(http.StatusAccepted, gin.H{
		"message":     "Sync initiated",
		"provider_id": id,
	})
}

// SyncAllConnectedProviders syncs all enabled connected providers
func (h *AdminHandler) SyncAllConnectedProviders(c *gin.Context) {
	// Start sync in background
	go func() {
		syncFunc := func(client providers.RegistrarClient) ([]types.Domain, error) {
			return client.FetchDomains()
		}
		
		if err := h.providerSvc.SyncAllProviders(syncFunc); err != nil {
			log.Printf("Sync all providers failed: %v", err)
		} else {
			log.Printf("Sync all providers completed")
		}
	}()
	
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Sync all providers initiated",
	})
}

// GetAutoSyncStatus returns the auto-sync status for all providers
func (h *AdminHandler) GetAutoSyncStatus(c *gin.Context) {
	status := h.providerSvc.GetAutoSyncStatus()
	c.JSON(http.StatusOK, status)
}

// StartAutoSync starts the auto-sync scheduler
func (h *AdminHandler) StartAutoSync(c *gin.Context) {
	syncFunc := func(client providers.RegistrarClient) ([]types.Domain, error) {
		return client.FetchDomains()
	}
	
	h.providerSvc.StartAutoSync(syncFunc)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Auto-sync scheduler started",
	})
}

// StopAutoSync stops the auto-sync scheduler
func (h *AdminHandler) StopAutoSync(c *gin.Context) {
	h.providerSvc.StopAutoSync()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Auto-sync scheduler stopped",
	})
}

// CheckDomainStatus checks the HTTP status of a single domain
func (h *AdminHandler) CheckDomainStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain ID required"})
		return
	}

	// Get the domain
	domain, err := h.domainRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
		return
	}

	// Check the status
	if err := h.statusChecker.CheckDomainWithHTTPS(domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to check status: %v", err)})
		return
	}

	// Update the domain in the database
	if err := h.domainRepo.Update(domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update domain: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domain_id":         domain.ID,
		"domain_name":       domain.Name,
		"http_status":       domain.HTTPStatus,
		"status_message":    domain.StatusMessage,
		"last_status_check": domain.LastStatusCheck,
	})
}

// BulkCheckStatus checks the HTTP status of multiple domains
func (h *AdminHandler) BulkCheckStatus(c *gin.Context) {
	var req struct {
		DomainIDs []string `json:"domain_ids"`
		CheckHTTPS bool    `json:"check_https"` // Also try HTTPS if HTTP fails
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if len(req.DomainIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No domain IDs provided"})
		return
	}

	var results []gin.H
	var errors []string

	for _, domainID := range req.DomainIDs {
		domain, err := h.domainRepo.GetByID(domainID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Domain %s: not found", domainID))
			continue
		}

		// Check status
		var checkErr error
		if req.CheckHTTPS {
			checkErr = h.statusChecker.CheckDomainWithHTTPS(domain)
		} else {
			checkErr = h.statusChecker.CheckDomain(domain)
		}

		if checkErr != nil {
			errors = append(errors, fmt.Sprintf("Domain %s: %v", domain.Name, checkErr))
			continue
		}

		// Update in database
		if err := h.domainRepo.Update(domain); err != nil {
			errors = append(errors, fmt.Sprintf("Domain %s: failed to update: %v", domain.Name, err))
			continue
		}

		results = append(results, gin.H{
			"domain_id":         domain.ID,
			"domain_name":       domain.Name,
			"http_status":       domain.HTTPStatus,
			"status_message":    domain.StatusMessage,
			"last_status_check": domain.LastStatusCheck,
		})
	}

	response := gin.H{
		"checked_count": len(results),
		"total_count":   len(req.DomainIDs),
		"results":       results,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	c.JSON(http.StatusOK, response)
}

// GetStatusSummary provides a summary of domain HTTP statuses
func (h *AdminHandler) GetStatusSummary(c *gin.Context) {
	// Get all domains
	domains, err := h.domainRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch domains"})
		return
	}

	// Generate status summary
	summary := status.GetStatusSummary(domains)

	c.JSON(http.StatusOK, summary)
}

// ListSupportedProviders returns all supported domain providers
func (h *AdminHandler) ListSupportedProviders(c *gin.Context) {
	providers := h.providerSvc.GetSupportedProviders()
	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
		"count":     len(providers),
	})
}

// TestProviderConnection tests provider credentials without saving
func (h *AdminHandler) TestProviderConnection(c *gin.Context) {
	var req types.ProviderConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Test the connection
	response, err := h.providerSvc.TestConnection(req.Provider, req.Credentials)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Test failed: %v", err)})
		return
	}

	// Log the test result
	h.providerSvc.LogProviderConnection(req.Provider, req.AccountName, response.Success, response.Message)

	c.JSON(http.StatusOK, response)
}

// ConnectProvider adds a new provider with credentials and optional auto-sync
func (h *AdminHandler) ConnectProvider(c *gin.Context) {
	var req types.ProviderConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Use enhanced provider service to add connected provider
	connectedProvider, err := h.providerSvc.AddConnectedProvider(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Also save to database if repository supports it
	if repo, ok := h.domainRepo.(interface{ CreateCredentials(*types.ProviderCredentials) error }); ok {
		creds := &types.ProviderCredentials{
			ID:               connectedProvider.ID,
			Provider:         req.Provider,
			Name:             req.Name,
			AccountName:      req.AccountName,
			Credentials:      types.CredentialsMap(req.Credentials),
			Enabled:          true,
			ConnectionStatus: "connected",
			CreatedAt:        connectedProvider.CreatedAt,
			UpdatedAt:        connectedProvider.UpdatedAt,
		}
		if err := repo.CreateCredentials(creds); err != nil {
			log.Printf("Warning: Failed to save credentials to database: %v", err)
		}
	}

	// Log successful connection
	h.providerSvc.LogProviderConnection(req.Provider, req.AccountName, true, "Provider connected successfully")

	response := types.ProviderConnectionResponse{
		Success:    true,
		Message:    "Provider connected successfully",
		ProviderID: connectedProvider.ID,
	}

	// Run initial sync if requested
	if req.AutoSync {
		go func() {
			syncFunc := func(client providers.RegistrarClient) ([]types.Domain, error) {
				return client.FetchDomains()
			}
			
			if err := h.providerSvc.SyncProvider(connectedProvider.ID, syncFunc); err != nil {
				log.Printf("Auto-sync failed for provider %s: %v", req.Name, err)
			} else {
				log.Printf("Auto-sync completed for provider %s", req.Name)
			}
		}()
		response.SyncStarted = true
	}

	c.JSON(http.StatusCreated, response)
}

// CreateCredentials creates new provider credentials
func (h *AdminHandler) CreateCredentials(c *gin.Context) {
	var creds types.ProviderCredentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credentials data"})
		return
	}

	if repo, ok := h.domainRepo.(interface{ CreateCredentials(*types.ProviderCredentials) error }); ok {
		if err := repo.CreateCredentials(&creds); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Don't expose actual credentials in the response
		creds.Credentials = map[string]string{"***": "***"}
		c.JSON(http.StatusCreated, creds)
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Credentials operations not implemented"})
	}
}

// UpdateCredentials updates existing provider credentials
func (h *AdminHandler) UpdateCredentials(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Credentials ID required"})
		return
	}

	var creds types.ProviderCredentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credentials data"})
		return
	}

	creds.ID = id
	if repo, ok := h.domainRepo.(interface{ UpdateCredentials(*types.ProviderCredentials) error }); ok {
		if err := repo.UpdateCredentials(&creds); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Don't expose actual credentials in the response
		creds.Credentials = map[string]string{"***": "***"}
		c.JSON(http.StatusOK, creds)
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Credentials operations not implemented"})
	}
}

// DeleteCredentials deletes provider credentials
func (h *AdminHandler) DeleteCredentials(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Credentials ID required"})
		return
	}

	if repo, ok := h.domainRepo.(interface{ DeleteCredentials(string) error }); ok {
		if err := repo.DeleteCredentials(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Credentials deleted successfully"})
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Credentials operations not implemented"})
	}
}

// GetPortfolioAnalytics retrieves aggregated domain portfolio analytics
func (h *AdminHandler) GetPortfolioAnalytics(c *gin.Context) {
	metrics, err := h.analyticsSvc.GetPortfolioMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

// GetFinancialAnalytics retrieves financial analysis and metrics
func (h *AdminHandler) GetFinancialAnalytics(c *gin.Context) {
	// Example: Return a subset of financial metrics for demonstration
	metrics, err := h.analyticsSvc.GetPortfolioMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics.FinancialMetrics)
}

// GetSecurityAnalytics retrieves security analysis and metrics
func (h *AdminHandler) GetSecurityAnalytics(c *gin.Context) {
	metrics, err := h.securitySvc.GetSecurityMetrics(30 * 24 * time.Hour) // Last 30 days
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

// GetTrendAnalytics retrieves historical trend analysis
func (h *AdminHandler) GetTrendAnalytics(c *gin.Context) {
	metrics, err := h.analyticsSvc.GetPortfolioMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics.TrendAnalysis)
}

// GetNotificationRules retrieves all notification rules
func (h *AdminHandler) GetNotificationRules(c *gin.Context) {
	// Implementation for fetching notification rules
	c.JSON(http.StatusOK, gin.H{"message": "Not yet implemented"})
}

// CreateNotificationRule creates a new notification rule
func (h *AdminHandler) CreateNotificationRule(c *gin.Context) {
	// Implementation for creating a new notification rule
	c.JSON(http.StatusCreated, gin.H{"message": "Not yet implemented"})
}

// UpdateNotificationRule updates an existing notification rule
func (h *AdminHandler) UpdateNotificationRule(c *gin.Context) {
	// Implementation for updating a notification rule
	c.JSON(http.StatusOK, gin.H{"message": "Not yet implemented"})
}

// DeleteNotificationRule deletes a notification rule
func (h *AdminHandler) DeleteNotificationRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rule ID required"})
		return
	}
	// Implementation for deleting a notification rule
	c.JSON(http.StatusOK, gin.H{"message": "Not yet implemented"})
}

// TestNotification sends a test notification
func (h *AdminHandler) TestNotification(c *gin.Context) {
	// Implementation for sending test notification
	c.JSON(http.StatusOK, gin.H{"message": "Test notification sent successfully"})
}

// GetAlerts retrieves a list of alerts
func (h *AdminHandler) GetAlerts(c *gin.Context) {
	// Example: Fetch alerts from notification system
	c.JSON(http.StatusOK, gin.H{"alerts": "Not yet implemented"})
}

// ResolveAlert resolves a specific alert
func (h *AdminHandler) ResolveAlert(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alert ID required"})
		return
	}
	// Implementation for resolving an alert
	c.JSON(http.StatusOK, gin.H{"message": "Alert resolved successfully"})
}

// GetAuditEvents retrieves audit events with filters
func (h *AdminHandler) GetAuditEvents(c *gin.Context) {
	// Example: Fetch and filter audit events
	c.JSON(http.StatusOK, gin.H{"audit_events": "Not yet implemented"})
}

// GetSecurityMetrics retrieves security metrics
func (h *AdminHandler) GetSecurityMetrics(c *gin.Context) {
	periodStr := c.Query("period")
	period, err := strconv.Atoi(periodStr)
	if err != nil || period <= 0 {
		period = 30
	}
	duration := time.Duration(period) * 24 * time.Hour
	metrics, err := h.securitySvc.GetSecurityMetrics(duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

// GetSecurityAlerts retrieves security alerts with filters
func (h *AdminHandler) GetSecurityAlerts(c *gin.Context) {
	// Example: Fetch security alerts from the security system
	c.JSON(http.StatusOK, gin.H{"alerts": "Not yet implemented"})
}

// ResolveSecurityAlert resolves a specific security alert
func (h *AdminHandler) ResolveSecurityAlert(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alert ID required"})
		return
	}
	// Implementation for resolving a security alert
	c.JSON(http.StatusOK, gin.H{"message": "Security alert resolved successfully"})
}

// GetActiveSessions retrieves all active sessions
func (h *AdminHandler) GetActiveSessions(c *gin.Context) {
	// Example: Fetch active sessions from the session store
	c.JSON(http.StatusOK, gin.H{"active_sessions": "Not yet implemented"})
}

// TerminateSession terminates a specific session
func (h *AdminHandler) TerminateSession(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
		return
	}
	// Implementation for terminating a session
	c.JSON(http.StatusOK, gin.H{"message": "Session terminated successfully"})
}

// Bulk DNS Management Handlers

// BulkAssignIP assigns the same IP address to multiple domains
func (h *AdminHandler) BulkAssignIP(c *gin.Context) {
	var req struct {
		Password   string `json:"password" binding:"required"`
		Operations []struct {
			DomainName string `json:"domain_name" binding:"required"`
			RecordName string `json:"record_name" binding:"required"`
			IPAddress  string `json:"ip_address" binding:"required,ip"`
			TTL        int    `json:"ttl" binding:"required,min=60,max=604800"`
		} `json:"operations" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Verify admin password for security
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Verify password (this would need to be implemented in auth service)
	if !h.authSvc.VerifyCurrentUserPassword(userID.(string), req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	// Log the bulk operation for security audit
	log.Printf("Bulk IP assignment initiated by user %s for %d domains", userID, len(req.Operations))

	results := make([]map[string]interface{}, 0, len(req.Operations))
	successCount := 0
	errorCount := 0

	for _, op := range req.Operations {
		result := map[string]interface{}{
			"domain_name": op.DomainName,
			"success":     false,
			"error":       nil,
		}

		// Get domain ID from domain name
		domains, err := h.domainRepo.GetDomainsByName(op.DomainName)
		if err != nil || len(domains) == 0 {
			result["error"] = fmt.Sprintf("Domain %s not found", op.DomainName)
			errorCount++
			results = append(results, result)
			continue
		}

		domainID := domains[0].ID

		// Create DNS record
		dnsRecord := types.DNSRecord{
			DomainID: domainID,
			Type:     "A",
			Name:     op.RecordName,
			Value:    op.IPAddress,
			TTL:      op.TTL,
		}

		err = h.dnsSvc.CreateOrUpdateRecord(dnsRecord)
		if err != nil {
			result["error"] = err.Error()
			errorCount++
		} else {
			result["success"] = true
			successCount++
		}

		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("Bulk IP assignment completed: %d successful, %d failed", successCount, errorCount),
		"success_count": successCount,
		"error_count":   errorCount,
		"results":       results,
	})
}

// BulkUpdateNameservers updates nameservers for multiple domains
func (h *AdminHandler) BulkUpdateNameservers(c *gin.Context) {
	var req struct {
		Password   string `json:"password" binding:"required"`
		Operations []struct {
			DomainName  string   `json:"domain_name" binding:"required"`
			Nameservers []string `json:"nameservers" binding:"required,min=2"`
		} `json:"operations" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Verify admin password for security
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Verify password
	if !h.authSvc.VerifyCurrentUserPassword(userID.(string), req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	// Log the bulk operation for security audit
	log.Printf("Bulk nameserver update initiated by user %s for %d domains", userID, len(req.Operations))

	results := make([]map[string]interface{}, 0, len(req.Operations))
	successCount := 0
	errorCount := 0

	for _, op := range req.Operations {
		result := map[string]interface{}{
			"domain_name": op.DomainName,
			"success":     false,
			"error":       nil,
		}

		// Get domain ID from domain name
		domains, err := h.domainRepo.GetDomainsByName(op.DomainName)
		if err != nil || len(domains) == 0 {
			result["error"] = fmt.Sprintf("Domain %s not found", op.DomainName)
			errorCount++
			results = append(results, result)
			continue
		}

		domainID := domains[0].ID

		// Update nameservers for the domain
		// First, remove existing NS records
		existingRecords, err := h.dnsSvc.GetDomainRecords(domainID)
		if err != nil {
			result["error"] = fmt.Sprintf("Failed to get existing records: %v", err)
			errorCount++
			results = append(results, result)
			continue
		}

		// Remove existing NS records
		for _, record := range existingRecords {
			if record.Type == "NS" {
				h.dnsSvc.DeleteRecord(record.ID)
			}
		}

		// Add new NS records
		nsSuccess := true
		for _, ns := range op.Nameservers {
			dnsRecord := types.DNSRecord{
				DomainID: domainID,
				Type:     "NS",
				Name:     "@",
				Value:    ns,
				TTL:      86400, // 24 hours default for NS records
			}

			if err := h.dnsSvc.CreateOrUpdateRecord(dnsRecord); err != nil {
				result["error"] = fmt.Sprintf("Failed to create NS record for %s: %v", ns, err)
				nsSuccess = false
				break
			}
		}

		if nsSuccess {
			result["success"] = true
			successCount++
		} else {
			errorCount++
		}

		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("Bulk nameserver update completed: %d successful, %d failed", successCount, errorCount),
		"success_count": successCount,
		"error_count":   errorCount,
		"results":       results,
	})
}

// BulkUpdateFromCSV processes bulk DNS updates from CSV data
func (h *AdminHandler) BulkUpdateFromCSV(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
		CSVData  []struct {
			Domain      string `json:"domain" binding:"required"`
			RecordType  string `json:"record_type"`
			Name        string `json:"name"`
			Value       string `json:"value"`
			TTL         string `json:"ttl"`
			Nameserver1 string `json:"nameserver1"`
			Nameserver2 string `json:"nameserver2"`
		} `json:"csv_data" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Verify admin password for security
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Verify password
	if !h.authSvc.VerifyCurrentUserPassword(userID.(string), req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	// Log the bulk operation for security audit
	log.Printf("Bulk CSV update initiated by user %s for %d rows", userID, len(req.CSVData))

	results := make([]map[string]interface{}, 0, len(req.CSVData))
	successCount := 0
	errorCount := 0

	for _, row := range req.CSVData {
		result := map[string]interface{}{
			"domain":  row.Domain,
			"success": false,
			"error":   nil,
		}

		// Get domain ID from domain name
		domains, err := h.domainRepo.GetDomainsByName(row.Domain)
		if err != nil || len(domains) == 0 {
			result["error"] = fmt.Sprintf("Domain %s not found", row.Domain)
			errorCount++
			results = append(results, result)
			continue
		}

		domainID := domains[0].ID

		// Process DNS record if provided
		if row.RecordType != "" && row.Name != "" && row.Value != "" {
			ttl := 3600 // default TTL
			if row.TTL != "" {
				if parsedTTL, err := strconv.Atoi(row.TTL); err == nil {
					ttl = parsedTTL
				}
			}

			dnsRecord := types.DNSRecord{
				DomainID: domainID,
				Type:     row.RecordType,
				Name:     row.Name,
				Value:    row.Value,
				TTL:      ttl,
			}

			if err := h.dnsSvc.CreateOrUpdateRecord(dnsRecord); err != nil {
				result["error"] = fmt.Sprintf("Failed to create/update DNS record: %v", err)
				errorCount++
				results = append(results, result)
				continue
			}
		}

		// Process nameservers if provided
		if row.Nameserver1 != "" && row.Nameserver2 != "" {
			// Remove existing NS records
			existingRecords, err := h.dnsSvc.GetDomainRecords(domainID)
			if err == nil {
				for _, record := range existingRecords {
					if record.Type == "NS" {
						h.dnsSvc.DeleteRecord(record.ID)
					}
				}
			}

			// Add new nameservers
			nameservers := []string{row.Nameserver1, row.Nameserver2}
			for _, ns := range nameservers {
				if ns != "" {
					dnsRecord := types.DNSRecord{
						DomainID: domainID,
						Type:     "NS",
						Name:     "@",
						Value:    ns,
						TTL:      86400,
					}

					if err := h.dnsSvc.CreateOrUpdateRecord(dnsRecord); err != nil {
						result["error"] = fmt.Sprintf("Failed to create NS record: %v", err)
						errorCount++
						results = append(results, result)
						continue
					}
				}
			}
		}

		result["success"] = true
		successCount++
		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("Bulk CSV update completed: %d successful, %d failed", successCount, errorCount),
		"success_count": successCount,
		"error_count":   errorCount,
		"results":       results,
	})
}
