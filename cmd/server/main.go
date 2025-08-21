package main

import (
	"github.com/joho/godotenv"
	"fmt"
	"log"
	"os"
	"strconv"
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
	"github.com/rusiqe/domainvault/internal/types"
	"github.com/rusiqe/domainvault/internal/uptimerobot"
)

func main() {
	// Load environment variables from .env file
	log.Printf("Current working directory: %s", func() string {
		if pwd, err := os.Getwd(); err == nil {
			return pwd
		}
		return "unknown"
	}())
	
	if err := godotenv.Overload(); err != nil {
		log.Printf("No .env file found or failed to load: %v", err)
	} else {
		log.Printf(".env file loaded successfully (with overrides)")
	}
	
	// Debug: Check if DATABASE_URL environment variable is set
	log.Printf("DATABASE_URL from env: %s", os.Getenv("DATABASE_URL"))
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}
	
	// Debug: print the loaded database URL
	log.Printf("Database URL: %s", cfg.DatabaseURL)

	// Initialize storage
	repo, err := storage.NewRepo(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer repo.Close()

	// Initialize sync service
	syncSvc := core.NewSyncService(repo)

	// Initialize providers
	providerSvc := providers.NewProviderService()
	for _, providerConfig := range cfg.Providers {
		client, err := providers.NewClient(providerConfig.Name, providerConfig.Credentials)
		if err != nil {
			log.Printf("Failed to initialize provider %s: %v", providerConfig.Name, err)
			continue
		}
		syncSvc.AddProvider(providerConfig.Name, client)
		providerSvc.RegisterClient(providerConfig.Name, client)
	}

	// Initialize DNS service early for schedulers
	dnsSvc := dns.NewDNSService(repo)
	// Configure sync service to use DNS service
	syncSvc.SetDNSService(dnsSvc)

	// Start domain sync scheduler (registrar domains)
	go func() {
		ticker := time.NewTicker(cfg.SyncInterval)
		defer ticker.Stop()

		for range ticker.C {
			if err := syncSvc.Run(); err != nil {
				log.Printf("Sync failed: %v", err)
			}
		}
	}()

	// Start DNS refresh scheduler (Cloudflare-first, then registrar as needed)
	go func() {
		intervalHours := 24
		if v := os.Getenv("DNS_SYNC_INTERVAL_HOURS"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				intervalHours = n
			}
		}
		dnsTicker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
		defer dnsTicker.Stop()

		// Run once on startup, then on each tick
		refresh := func() {
			cfClient, ok := providerSvc.GetClientByProviderName("cloudflare")
			if !ok {
				log.Printf("DNS refresh: Cloudflare client not available; skipping")
				return
			}
			domains, err := repo.GetAll()
			if err != nil {
				log.Printf("DNS refresh: failed to list domains: %v", err)
				return
			}
			updated := 0
			skipped := 0
			for _, d := range domains {
				// Fetch from Cloudflare
				records, err := cfClient.FetchDNSRecords(d.Name)
				if err != nil {
					// Optional: try registrar as fallback
					if regClient, ok := providerSvc.GetClientByProviderName(d.Provider); ok {
						records, err = regClient.FetchDNSRecords(d.Name)
					}
				}
				if err != nil || len(records) == 0 {
					continue
				}
				// Compare with DB
				stored, err := dnsSvc.GetDomainRecords(d.ID)
				if err != nil {
					log.Printf("DNS refresh: failed to get stored records for %s: %v", d.Name, err)
				}
				if dnsRecordSetsEqual(stored, records) {
					skipped++
					continue
				}
				// Replace stored records with fresh ones
				if err := dnsSvc.BulkUpdateRecords(d.ID, normalizeRecordsForStore(d.ID, records)); err != nil {
					log.Printf("DNS refresh: failed to update records for %s: %v", d.Name, err)
					continue
				}
				updated++
			}
			log.Printf("DNS refresh completed: %d updated, %d unchanged, total %d", updated, skipped, len(domains))
		}

		// Initial run
		refresh()
		for range dnsTicker.C {
			refresh()
		}
	}()

	// Initialize services
	authSvc := auth.NewAuthService(repo)

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

// Initialize UptimeRobot service
var uptimeRobotSvc *uptimerobot.Service
if cfg.UptimeRobot != nil {
	uptimeRobotSvc = uptimerobot.NewService(cfg.UptimeRobot)
	if uptimeRobotSvc.IsConfigured() {
		log.Printf("UptimeRobot service initialized")
	} else {
		log.Printf("UptimeRobot service disabled: not properly configured")
	}
} else {
	log.Printf("UptimeRobot service disabled: no configuration found")
	// Create a nil service for graceful handling
	uptimeRobotSvc = nil
}

// Initialize API handlers (with UptimeRobot service)
handler := api.NewDomainHandler(repo, syncSvc, uptimeRobotSvc)
adminHandler := api.NewAdminHandler(repo, authSvc, syncSvc, dnsSvc, providerSvc, analyticsSvc, notificationSvc, securitySvc, uptimeRobotSvc)

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

// dnsRecordSetsEqual compares two DNS record sets for equality ignoring IDs and timestamps.
func dnsRecordSetsEqual(a, b []types.DNSRecord) bool {
	if len(a) != len(b) {
		return false
	}
	am := make(map[string]int, len(a))
	for _, r := range a {
		am[fingerprint(r)]++
	}
	for _, r := range b {
		key := fingerprint(r)
		if c, ok := am[key]; !ok || c == 0 {
			return false
		} else {
			am[key] = c - 1
		}
	}
	for _, c := range am {
		if c != 0 {
			return false
		}
	}
	return true
}

func fingerprint(r types.DNSRecord) string {
	prio := ""
	if r.Priority != nil {
		prio = fmt.Sprintf("|p=%d", *r.Priority)
	}
	return fmt.Sprintf("%s|%s|%s|ttl=%d%s", r.Type, r.Name, r.Value, r.TTL, prio)
}

// normalizeRecordsForStore sets DomainID and timestamps; values are already normalized by provider clients
func normalizeRecordsForStore(domainID string, in []types.DNSRecord) []types.DNSRecord {
	out := make([]types.DNSRecord, len(in))
	now := time.Now()
	for i := range in {
		out[i] = in[i]
		out[i].DomainID = domainID
		// Keep TTL/priority as provided
		out[i].CreatedAt = now
		out[i].UpdatedAt = now
	}
	return out
}
