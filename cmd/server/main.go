package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusiqe/domainvault/internal/api"
	"github.com/rusiqe/domainvault/internal/config"
	"github.com/rusiqe/domainvault/internal/core"
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

	// Initialize API handlers
	handler := api.NewDomainHandler(repo, syncSvc)

	// Setup Gin router
	r := gin.Default()
	handler.RegisterRoutes(r)

	// Start server
	log.Printf("Starting server on port %d", cfg.Port)
	if err := r.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
