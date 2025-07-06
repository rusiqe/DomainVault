package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusiqe/domainvault/internal/core"
	"github.com/rusiqe/domainvault/internal/storage"
	"github.com/rusiqe/domainvault/internal/types"
)

// DomainHandler handles HTTP requests for domain operations
type DomainHandler struct {
	repo    storage.DomainRepository
	syncSvc *core.SyncService
}

// NewDomainHandler creates a new domain handler
func NewDomainHandler(repo storage.DomainRepository, syncSvc *core.SyncService) *DomainHandler {
	return &DomainHandler{
		repo:    repo,
		syncSvc: syncSvc,
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

// DeleteDomain removes a domain by ID
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

	c.JSON(http.StatusOK, gin.H{"message": "domain deleted successfully"})
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

// TriggerSync starts a full sync across all providers
func (h *DomainHandler) TriggerSync(c *gin.Context) {
	go func() {
		if err := h.syncSvc.Run(); err != nil {
			// Log error, but don't block the response
			// In production, this would be logged properly
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "sync started",
		"status":  "accepted",
	})
}

// SyncProvider triggers sync for a specific provider
func (h *DomainHandler) SyncProvider(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider name required"})
		return
	}

	go func() {
		if err := h.syncSvc.SyncProvider(provider); err != nil {
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
