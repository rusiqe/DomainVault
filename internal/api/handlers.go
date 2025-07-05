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
		api.DELETE("/domains/:id", h.DeleteDomain)
		api.GET("/domains/summary", h.GetSummary)
		api.GET("/domains/expiring", h.GetExpiringDomains)

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
