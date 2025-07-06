package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rusiqe/domainvault/internal/auth"
	"github.com/rusiqe/domainvault/internal/core"
	"github.com/rusiqe/domainvault/internal/dns"
	"github.com/rusiqe/domainvault/internal/status"
	"github.com/rusiqe/domainvault/internal/storage"
	"github.com/rusiqe/domainvault/internal/types"
)

// AdminHandler handles admin-specific HTTP requests
type AdminHandler struct {
	domainRepo    storage.DomainRepository
	authSvc       *auth.AuthService
	syncSvc       *core.SyncService
	dnsSvc        *dns.DNSService
	statusChecker *status.StatusChecker
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	domainRepo storage.DomainRepository,
	authSvc *auth.AuthService,
	syncSvc *core.SyncService,
	dnsSvc *dns.DNSService,
) *AdminHandler {
	return &AdminHandler{
		domainRepo:    domainRepo,
		authSvc:       authSvc,
		syncSvc:       syncSvc,
		dnsSvc:        dnsSvc,
		statusChecker: status.NewStatusChecker(),
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
