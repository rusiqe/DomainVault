package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusiqe/domainvault/internal/analytics"
	"github.com/rusiqe/domainvault/internal/api"
	"github.com/rusiqe/domainvault/internal/auth"
	"github.com/rusiqe/domainvault/internal/config"
	"github.com/rusiqe/domainvault/internal/core"
	"github.com/rusiqe/domainvault/internal/dns"
	"github.com/rusiqe/domainvault/internal/notifications"
	"github.com/rusiqe/domainvault/internal/providers"
	"github.com/rusiqe/domainvault/internal/security"
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

	// Initialize enhanced services
	analyticsSvc := analytics.NewAnalyticsService(repo)

	// Initialize notification service with default configuration
	emailConfig := notifications.EmailConfig{
		SMTPHost:    "smtp.gmail.com",
		SMTPPort:    587,
		FromAddress: "noreply@domainvault.com",
		FromName:    "DomainVault",
		Enabled:     false, // Disabled by default
	}
	slackConfig := notifications.SlackConfig{
		Enabled: false, // Disabled by default
	}
	webhookConfig := notifications.WebhookConfig{
		Enabled: false, // Disabled by default
	}
	notificationSvc := notifications.NewNotificationService(emailConfig, slackConfig, webhookConfig)

	// Initialize security service with default configuration
	securityConfig := security.SecurityConfig{
		MaxLoginAttempts:     5,
		LockoutDuration:      15 * time.Minute,
		SessionTimeout:       24 * time.Hour,
		PasswordMinLength:    8,
		RequireStrongPassword: true,
		EnableAuditLogging:   true,
		EnableIPWhitelisting: false,
		EnableMFA:           false,
		JWTSigningKey:       "your-secret-key-change-in-production",
	}
	// Note: In production, these would be implemented with actual repository interfaces
	securitySvc := security.NewSecurityService(nil, nil, nil, securityConfig)

	// Create default admin user if it doesn't exist
	if err := authSvc.CreateDefaultAdmin(); err != nil {
		log.Printf("Warning: Failed to create default admin user: %v", err)
	}

	// Initialize API handlers
	handler := api.NewDomainHandler(repo, syncSvc)
	adminHandler := api.NewAdminHandler(repo, authSvc, syncSvc, dnsSvc, analyticsSvc, notificationSvc, securitySvc)

	// Setup Gin router
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./web/static")

	// Serve admin interface
	r.GET("/admin", func(c *gin.Context) {
		c.File("./web/admin.html")
	})

	// Serve enhanced admin interface
	r.GET("/admin-enhanced.html", func(c *gin.Context) {
		c.File("./web/admin-enhanced.html")
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
