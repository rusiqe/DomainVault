package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusiqe/domainvault/internal/api"
	"github.com/rusiqe/domainvault/internal/auth"
	"github.com/rusiqe/domainvault/internal/config"
	"github.com/rusiqe/domainvault/internal/core"
	"github.com/rusiqe/domainvault/internal/dns"
	"github.com/rusiqe/domainvault/internal/providers"
	"github.com/rusiqe/domainvault/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize storage
	repo, err := storage.NewPostgresRepo(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer repo.Close()

	// Initialize sync service
	syncSvc := core.NewSyncService(repo)

	// Initialize providers
	for _, providerConfig := range cfg.Providers {
		client, err := providers.NewClient(providerConfig.Name, providerConfig.Credentials)
		if err != nil {
			log.Printf("Failed to initialize provider %s: %v", providerConfig.Name, err)
			continue
		}
		syncSvc.AddProvider(providerConfig.Name, client)
	}

	// Start sync scheduler
	go func() {
		ticker := time.NewTicker(cfg.SyncInterval)
		defer ticker.Stop()

		for range ticker.C {
			if err := syncSvc.Run(); err != nil {
				log.Printf("Sync failed: %v", err)
			}
		}
	}()

	// Initialize services
	authSvc := auth.NewAuthService(repo)
	dnsSvc := dns.NewDNSService(repo)

	// Create default admin user if it doesn't exist
	if err := authSvc.CreateDefaultAdmin(); err != nil {
		log.Printf("Warning: Failed to create default admin user: %v", err)
	}

	// Initialize API handlers
	handler := api.NewDomainHandler(repo, syncSvc)
	adminHandler := api.NewAdminHandler(repo, authSvc, syncSvc, dnsSvc)

	// Setup Gin router
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./web/static")

	// Serve admin interface
	r.GET("/admin", func(c *gin.Context) {
		c.File("./web/admin.html")
	})

	// Serve main interface
	r.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	// Register API routes
	handler.RegisterRoutes(r)
	adminHandler.RegisterAdminRoutes(r)

	// Start server
	log.Printf("Starting server on port %d", cfg.Port)
	log.Printf("Admin interface available at: http://localhost:%d/admin", cfg.Port)
	log.Printf("Default admin credentials: admin / admin123")
	if err := r.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
