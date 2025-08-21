package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusiqe/domainvault/internal/core"
	"github.com/rusiqe/domainvault/internal/storage"
	"github.com/rusiqe/domainvault/internal/types"
	"github.com/rusiqe/domainvault/internal/uptimerobot"
)

// DomainHandler handles HTTP requests for domain operations
type DomainHandler struct {
	repo      storage.DomainRepository
	syncSvc   *core.SyncService
	uptimeSvc *uptimerobot.Service
}

// NewDomainHandler creates a new domain handler
func NewDomainHandler(repo storage.DomainRepository, syncSvc *core.SyncService, uptimeSvc *uptimerobot.Service) *DomainHandler {
	return &DomainHandler{
		repo:      repo,
		syncSvc:   syncSvc,
		uptimeSvc: uptimeSvc,
	}
}

// RegisterRoutes sets up the HTTP routes
func (h *DomainHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		// Domain operations
		api.GET("/domains", h.ListDomains)
		api.GET("/domains/:id", h.GetDomain)
		api.PUT("/domains/:id", h.UpdateDomain)
api.DELETE("/domains/:id", h.DeleteDomain)
		api.PUT("/domains/:id/visibility", h.SetDomainVisibility)
		api.GET("/domains/summary", h.GetSummary)
		api.GET("/domains/expiring", h.GetExpiringDomains)

		// Category operations
		api.GET("/categories", h.ListCategories)
		api.POST("/categories", h.CreateCategory)
		api.GET("/categories/:id", h.GetCategory)
		api.PUT("/categories/:id", h.UpdateCategory)
		api.DELETE("/categories/:id", h.DeleteCategory)

		// Project operations
		api.GET("/projects", h.ListProjects)
		api.POST("/projects", h.CreateProject)
		api.GET("/projects/:id", h.GetProject)
		api.PUT("/projects/:id", h.UpdateProject)
		api.DELETE("/projects/:id", h.DeleteProject)

		// Provider credentials operations
		api.GET("/credentials", h.ListCredentials)
		api.POST("/credentials", h.CreateCredentials)
		api.GET("/credentials/:id", h.GetCredentials)
		api.PUT("/credentials/:id", h.UpdateCredentials)
		api.DELETE("/credentials/:id", h.DeleteCredentials)

		// Import operations
		api.POST("/import", h.ImportDomains)
		api.GET("/providers", h.ListProviders)

		// Sync operations
		api.POST("/sync", h.TriggerSync)
		api.POST("/sync/:provider", h.SyncProvider)
		api.GET("/sync/status", h.GetSyncStatus)

		// UptimeRobot monitoring
		api.POST("/monitoring/sync", h.SyncMonitoring)
		api.POST("/monitoring/create", h.CreateMonitoring)
		api.GET("/monitoring/stats", h.GetMonitoringStats)

		// Health check
		api.GET("/health", h.HealthCheck)
	}
}

// ListDomains returns a list of domains with optional filtering
func (h *DomainHandler) ListDomains(c *gin.Context) {
	filter := types.DomainFilter{}

	// Parse query parameters
	if provider := c.Query("provider"); provider != "" {
		filter.Provider = provider
	}

	if search := c.Query("search"); search != "" {
		filter.Search = search
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	} else {
		filter.Limit = 50 // Default limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Parse date filters
	if expiresAfter := c.Query("expires_after"); expiresAfter != "" {
		if date, err := time.Parse("2006-01-02", expiresAfter); err == nil {
			filter.ExpiresAfter = &date
		}
	}

	if expiresBefore := c.Query("expires_before"); expiresBefore != "" {
		if date, err := time.Parse("2006-01-02", expiresBefore); err == nil {
			filter.ExpiresBefore = &date
		}
	}

	if c.Query("include_hidden") == "true" {
		filter.IncludeHidden = true
	}
	if c.Query("only_hidden") == "true" {
		filter.OnlyHidden = true
		filter.IncludeHidden = true
	}

	domains, err := h.repo.GetByFilter(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domains": domains,
		"count":   len(domains),
		"filter":  filter,
	})
}

// GetDomain returns a specific domain by ID
func (h *DomainHandler) GetDomain(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain ID required"})
		return
	}

	domain, err := h.repo.GetByID(id)
	if err != nil {
		if err == types.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, domain)
}

// DeleteDomain removes a domain by ID (soft delete: sets visible=false)
func (h *DomainHandler) DeleteDomain(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain ID required"})
		return
	}

	err := h.repo.Delete(id)
	if err != nil {
		if err == types.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "domain removed from portfolio"})
}

// SetDomainVisibility toggles a domain's visibility (soft delete/restore)
func (h *DomainHandler) SetDomainVisibility(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain ID required"})
		return
	}
	var req struct{ Visible bool `json:"visible"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.repo.SetVisibility(id, req.Visible); err != nil {
		if err == types.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	status := "hidden"
	if req.Visible {
		status = "visible"
	}
	c.JSON(http.StatusOK, gin.H{"message": "domain visibility updated", "status": status})
}

// GetSummary returns domain statistics
func (h *DomainHandler) GetSummary(c *gin.Context) {
	summary, err := h.repo.GetSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetExpiringDomains returns domains expiring within a threshold
func (h *DomainHandler) GetExpiringDomains(c *gin.Context) {
	// Default to 30 days
	threshold := 30 * 24 * time.Hour

	if daysStr := c.Query("days"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 {
			threshold = time.Duration(days) * 24 * time.Hour
		}
	}

	domains, err := h.repo.GetExpiring(threshold)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domains":   domains,
		"count":     len(domains),
		"threshold": threshold.String(),
	})
}

// TriggerSync starts a full sync across all providers including DNS records
func (h *DomainHandler) TriggerSync(c *gin.Context) {
	go func() {
		if err := h.syncSvc.SyncDomainsWithDNS(); err != nil {
			// Log error, but don't block the response
			// In production, this would be logged properly
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "sync started",
		"status":  "accepted",
	})
}

// SyncProvider triggers sync for a specific provider including DNS records
func (h *DomainHandler) SyncProvider(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider name required"})
		return
	}

	go func() {
		if err := h.syncSvc.SyncProviderWithDNS(provider); err != nil {
			// Log error, but don't block the response
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message":  "provider sync started",
		"provider": provider,
		"status":   "accepted",
	})
}

// GetSyncStatus returns the current sync service status
func (h *DomainHandler) GetSyncStatus(c *gin.Context) {
	status := h.syncSvc.GetStatus()
	c.JSON(http.StatusOK, status)
}

// HealthCheck returns the service health status
func (h *DomainHandler) HealthCheck(c *gin.Context) {
	// Check database connection
	if err := h.repo.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "unhealthy",
			"database": "disconnected",
			"error":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"database": "connected",
		"version":  "1.0.0",
	})
}

// ============================================================================
// DOMAIN UPDATE METHOD
// ============================================================================

// UpdateDomain updates domain details
func (h *DomainHandler) UpdateDomain(c *gin.Context) {
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
	if err := h.repo.Update(&domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain)
}

// ============================================================================
// CATEGORY MANAGEMENT METHODS
// ============================================================================

// ListCategories returns all categories
func (h *DomainHandler) ListCategories(c *gin.Context) {
	if repo, ok := h.repo.(interface{ GetAllCategories() ([]types.Category, error) }); ok {
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
func (h *DomainHandler) CreateCategory(c *gin.Context) {
	var category types.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category data"})
		return
	}

	if repo, ok := h.repo.(interface{ CreateCategory(*types.Category) error }); ok {
		if err := repo.CreateCategory(&category); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, category)
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Category operations not implemented"})
	}
}

// GetCategory returns a specific category by ID
func (h *DomainHandler) GetCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID required"})
		return
	}

	if repo, ok := h.repo.(interface{ GetCategoryByID(string) (*types.Category, error) }); ok {
		category, err := repo.GetCategoryByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
		c.JSON(http.StatusOK, category)
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Category operations not implemented"})
	}
}

// UpdateCategory updates an existing category
func (h *DomainHandler) UpdateCategory(c *gin.Context) {
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
	if repo, ok := h.repo.(interface{ UpdateCategory(*types.Category) error }); ok {
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
func (h *DomainHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID required"})
		return
	}

	if repo, ok := h.repo.(interface{ DeleteCategory(string) error }); ok {
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
func (h *DomainHandler) ListProjects(c *gin.Context) {
	if repo, ok := h.repo.(interface{ GetAllProjects() ([]types.Project, error) }); ok {
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
func (h *DomainHandler) CreateProject(c *gin.Context) {
	var project types.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project data"})
		return
	}

	if repo, ok := h.repo.(interface{ CreateProject(*types.Project) error }); ok {
		if err := repo.CreateProject(&project); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, project)
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Project operations not implemented"})
	}
}

// GetProject returns a specific project by ID
func (h *DomainHandler) GetProject(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID required"})
		return
	}

	if repo, ok := h.repo.(interface{ GetProjectByID(string) (*types.Project, error) }); ok {
		project, err := repo.GetProjectByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusOK, project)
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Project operations not implemented"})
	}
}

// UpdateProject updates an existing project
func (h *DomainHandler) UpdateProject(c *gin.Context) {
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
	if repo, ok := h.repo.(interface{ UpdateProject(*types.Project) error }); ok {
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
func (h *DomainHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID required"})
		return
	}

	if repo, ok := h.repo.(interface{ DeleteProject(string) error }); ok {
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
func (h *DomainHandler) ListCredentials(c *gin.Context) {
	if repo, ok := h.repo.(interface{ GetAllCredentials() ([]types.ProviderCredentials, error) }); ok {
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

// CreateCredentials creates new provider credentials
func (h *DomainHandler) CreateCredentials(c *gin.Context) {
	var creds types.ProviderCredentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credentials data"})
		return
	}

	if repo, ok := h.repo.(interface{ CreateCredentials(*types.ProviderCredentials) error }); ok {
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

// GetCredentials returns specific provider credentials by ID
func (h *DomainHandler) GetCredentials(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Credentials ID required"})
		return
	}

	if repo, ok := h.repo.(interface{ GetCredentialsByID(string) (*types.ProviderCredentials, error) }); ok {
		creds, err := repo.GetCredentialsByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Credentials not found"})
			return
		}

		// Don't expose actual credentials in the response
		creds.Credentials = map[string]string{"***": "***"}
		c.JSON(http.StatusOK, creds)
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Credentials operations not implemented"})
	}
}

// UpdateCredentials updates existing provider credentials
func (h *DomainHandler) UpdateCredentials(c *gin.Context) {
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
	if repo, ok := h.repo.(interface{ UpdateCredentials(*types.ProviderCredentials) error }); ok {
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
func (h *DomainHandler) DeleteCredentials(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Credentials ID required"})
		return
	}

	if repo, ok := h.repo.(interface{ DeleteCredentials(string) error }); ok {
		if err := repo.DeleteCredentials(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Credentials deleted successfully"})
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Credentials operations not implemented"})
	}
}

// ============================================================================
// ADDITIONAL METHODS
// ============================================================================

// ImportDomains handles domain import operations
func (h *DomainHandler) ImportDomains(c *gin.Context) {
	var req types.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid import request"})
		return
	}

	// This would trigger import from the specified provider
	// For now, return a placeholder response
	c.JSON(http.StatusAccepted, gin.H{
		"message":  "Import initiated",
		"provider": req.Provider,
		"status":   "pending",
	})
}

// ListProviders returns available domain providers
func (h *DomainHandler) ListProviders(c *gin.Context) {
	providers := []string{"mock", "godaddy", "namecheap", "cloudflare"}
	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
		"count":     len(providers),
	})
}

// ============================================================================
// UPTIMEROBOT MONITORING METHODS
// ============================================================================

// SyncMonitoring synchronizes UptimeRobot monitoring for domains
func (h *DomainHandler) SyncMonitoring(c *gin.Context) {
	// Since we're using Terraform for UptimeRobot management,
	// return information about Terraform-managed monitoring
	c.JSON(http.StatusOK, gin.H{
		"message": "UptimeRobot monitoring is managed via Terraform",
		"status": "terraform_managed",
		"info": "Use 'terraform apply' in the terraform/ directory to sync monitors",
		"terraform_dir": "./terraform/",
		"managed_domains": []string{
			"content-dao.xyz",
			"atozpolandmoves.com", 
			"zurioasiselite.icu",
			"relationswithdevs.xyz",
		},
	})
}

// CreateMonitoring creates UptimeRobot monitors for specified domains
func (h *DomainHandler) CreateMonitoring(c *gin.Context) {
	// Since we're using Terraform for UptimeRobot management,
	// redirect users to use Terraform instead
	c.JSON(http.StatusOK, gin.H{
		"message": "Monitor creation is managed via Terraform",
		"status": "terraform_managed",
		"instructions": "Add domains to terraform/terraform.tfvars and run 'terraform apply'",
		"terraform_dir": "./terraform/",
		"current_monitors": []string{
			"content-dao.xyz (ID: 801092250)",
			"atozpolandmoves.com (ID: 801092248)",
			"zurioasiselite.icu (ID: 801092247)",
			"relationswithdevs.xyz (ID: 801092249)",
		},
	})
}

// GetMonitoringStats returns UptimeRobot monitoring statistics
func (h *DomainHandler) GetMonitoringStats(c *gin.Context) {
	// Check if UptimeRobot is configured
	if h.uptimeSvc == nil || !h.uptimeSvc.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "UptimeRobot is not configured",
			"message": "Please configure UptimeRobot API key first",
		})
		return
	}

	// Get live monitoring metrics from UptimeRobot
	uptimeMetrics, err := h.uptimeSvc.GetMonitoringMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get UptimeRobot monitoring statistics",
			"details": err.Error(),
		})
		return
	}

	// Get live DomainVault monitors with detailed stats
	liveMonitors, err := h.uptimeSvc.GetDomainVaultMonitors(true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get live monitor data",
			"details": err.Error(),
		})
		return
	}

	// Get database monitoring stats
	dbStats := h.getDBMonitoringStats()

	// Create detailed monitor breakdown
	monitorDetails := make([]map[string]interface{}, 0, len(liveMonitors))
	for _, monitor := range liveMonitors {
		detail := map[string]interface{}{
			"id": monitor.ID,
			"name": monitor.FriendlyName,
			"url": monitor.URL,
			"status": h.getStatusString(monitor.Status),
			"type": h.getMonitorTypeString(monitor.Type),
			"interval": monitor.Interval,
			"created_at": time.Unix(monitor.CreateDatetime, 0).Format(time.RFC3339),
		}

		// Add response time if available
		if len(monitor.ResponseTimes) > 0 {
			detail["response_time_ms"] = monitor.ResponseTimes[len(monitor.ResponseTimes)-1].Value
		}

		// Add recent logs if available
		if len(monitor.Logs) > 0 {
			recentLogs := make([]map[string]interface{}, 0, len(monitor.Logs))
			for _, log := range monitor.Logs {
				recentLogs = append(recentLogs, map[string]interface{}{
					"datetime": time.Unix(log.Datetime, 0).Format(time.RFC3339),
					"type": log.Type,
					"duration": log.Duration,
					"reason": log.Reason.Detail,
				})
			}
			detail["recent_events"] = recentLogs
		}

		monitorDetails = append(monitorDetails, detail)
	}

	// Calculate summary statistics
	summary := map[string]interface{}{
		"total_monitors": len(liveMonitors),
		"up_monitors": 0,
		"down_monitors": 0,
		"paused_monitors": 0,
		"average_response_time": 0,
		"monitored_domains": len(liveMonitors),
	}

	totalResponseTime := 0
	responseTimeCount := 0

	for _, monitor := range liveMonitors {
		switch monitor.Status {
		case 2: // Up
			summary["up_monitors"] = summary["up_monitors"].(int) + 1
		case 8, 9: // Seems down or down
			summary["down_monitors"] = summary["down_monitors"].(int) + 1
		case 0: // Paused
			summary["paused_monitors"] = summary["paused_monitors"].(int) + 1
		}

		// Calculate average response time
		if len(monitor.ResponseTimes) > 0 {
			totalResponseTime += monitor.ResponseTimes[len(monitor.ResponseTimes)-1].Value
			responseTimeCount++
		}
	}

	if responseTimeCount > 0 {
		summary["average_response_time"] = totalResponseTime / responseTimeCount
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"timestamp": time.Now().Format(time.RFC3339),
		"summary": summary,
		"uptimerobot_account": uptimeMetrics["account"],
		"database_stats": dbStats,
		"monitor_details": monitorDetails,
		"last_updated": time.Now().Format(time.RFC3339),
	})
}

// getDBMonitoringStats retrieves monitoring statistics from the database
func (h *DomainHandler) getDBMonitoringStats() map[string]interface{} {
	// Get domain summary to provide real database stats
	summary, err := h.repo.GetSummary()
	if err != nil {
		// Return empty stats if we can't get summary
		return map[string]interface{}{
			"total_domains": 0,
			"monitored_domains": 0,
			"up_domains": 0,
			"down_domains": 0,
			"average_uptime": 0.0,
			"average_response_time": 0,
			"error": "Failed to get database stats",
		}
	}

	return map[string]interface{}{
		"total_domains": summary.Total,
		"domains_by_provider": summary.ByProvider,
		"expiring_soon": summary.ExpiringIn,
		"last_sync": summary.LastSync,
		"monitored_domains": 0, // This would be populated from actual monitoring data
		"up_domains": 0,
		"down_domains": 0,
		"average_uptime": 0.0,
		"average_response_time": 0,
	}
}

// getStatusString converts monitor status to human-readable string
func (h *DomainHandler) getStatusString(status uptimerobot.MonitorStatus) string {
	switch status {
	case 0:
		return "paused"
	case 1:
		return "not_checked_yet"
	case 2:
		return "up"
	case 8:
		return "seems_down"
	case 9:
		return "down"
	default:
		return "unknown"
	}
}

// getMonitorTypeString converts monitor type to human-readable string
func (h *DomainHandler) getMonitorTypeString(monitorType uptimerobot.MonitorType) string {
	switch monitorType {
	case 1:
		return "http"
	case 2:
		return "keyword"
	case 3:
		return "ping"
	case 4:
		return "port"
	case 5:
		return "heartbeat"
	default:
		return "unknown"
	}
}
