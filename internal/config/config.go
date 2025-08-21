package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rusiqe/domainvault/internal/types"
)

// Config holds all application configuration
type Config struct {
	Port         int                    `json:"port"`
	DatabaseURL  string                 `json:"database_url"`
	SyncInterval time.Duration          `json:"sync_interval"`
	LogLevel     string                 `json:"log_level"`
	Providers    []ProviderConfig       `json:"providers"`
	UptimeRobot  *UptimeRobotConfig     `json:"uptime_robot,omitempty"`
}

// ProviderConfig holds configuration for a domain registrar
type ProviderConfig struct {
	Name        string                 `json:"name"`
	Enabled     bool                   `json:"enabled"`
	Credentials map[string]interface{} `json:"credentials"`
}

// UptimeRobotConfig holds UptimeRobot monitoring configuration
type UptimeRobotConfig struct {
	APIKey             string   `json:"api_key"`
	Enabled            bool     `json:"enabled"`
	Interval           int      `json:"interval"`              // Check interval in seconds
	AlertContacts      []string `json:"alert_contacts"`        // Alert contact IDs
	AutoCreateMonitors bool     `json:"auto_create_monitors"`  // Auto-create monitors for new domains
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Port:         getEnvInt("PORT", 8080),
		DatabaseURL:  getEnvString("DATABASE_URL", "postgres://user:password@localhost/domainvault?sslmode=disable"),
		SyncInterval: getEnvDuration("SYNC_INTERVAL", "1h"),
		LogLevel:     getEnvString("LOG_LEVEL", "info"),
		Providers:    loadProviders(),
		UptimeRobot:  loadUptimeRobotConfig(),
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate ensures configuration is valid
func (c *Config) validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return types.ErrInvalidConfig
	}
	if c.DatabaseURL == "" {
		return types.ErrMissingConfig
	}
	if c.SyncInterval < time.Minute {
		return types.ErrInvalidConfig
	}
	return nil
}

// loadProviders loads provider configurations from environment
func loadProviders() []ProviderConfig {
	var providers []ProviderConfig

	// GoDaddy configuration
	if godaddyKey := getEnvString("GODADDY_API_KEY", ""); godaddyKey != "" {
		providers = append(providers, ProviderConfig{
			Name:    "godaddy",
			Enabled: true,
			Credentials: map[string]interface{}{
				"api_key": godaddyKey,
				"api_secret": getEnvString("GODADDY_API_SECRET", ""),
			},
		})
	}

	// Namecheap configuration
	if namecheapKey := getEnvString("NAMECHEAP_API_KEY", ""); namecheapKey != "" {
		providers = append(providers, ProviderConfig{
			Name:    "namecheap",
			Enabled: true,
			Credentials: map[string]interface{}{
				"api_key": namecheapKey,
				"username": getEnvString("NAMECHEAP_USERNAME", ""),
			},
		})
	}

	// Hostinger configuration
	if hostingerKey := getEnvString("HOSTINGER_API_KEY", ""); hostingerKey != "" {
		providers = append(providers, ProviderConfig{
			Name:    "hostinger",
			Enabled: true,
			Credentials: map[string]interface{}{
				"api_key": hostingerKey,
			},
		})
	}

	// Cloudflare DNS configuration (DNS-only provider)
	if cfToken := getEnvString("CLOUDFLARE_API_TOKEN", ""); cfToken != "" {
		providers = append(providers, ProviderConfig{
			Name:    "cloudflare",
			Enabled: true,
			Credentials: map[string]interface{}{
				"api_token": cfToken,
			},
		})
	}

	// Mock provider configuration (for testing and demo purposes)
	if getEnvBool("ENABLE_MOCK_PROVIDER", false) {
		providers = append(providers, ProviderConfig{
			Name:    "mock",
			Enabled: true,
			Credentials: map[string]interface{}{
				"mock": true, // Mock provider doesn't need real credentials
			},
		})
	}

	return providers
}

// loadUptimeRobotConfig loads UptimeRobot configuration from environment
func loadUptimeRobotConfig() *UptimeRobotConfig {
	apiKey := getEnvString("UPTIMEROBOT_API_KEY", "")
	if apiKey == "" {
		return nil // UptimeRobot not configured
	}

	// Parse alert contacts from comma-separated string
	var alertContacts []string
	if contactsStr := getEnvString("UPTIMEROBOT_ALERT_CONTACTS", ""); contactsStr != "" {
		// Simple split by comma - in production you might want more robust parsing
		for _, contact := range splitString(contactsStr, ",") {
			if contact := trimString(contact); contact != "" {
				alertContacts = append(alertContacts, contact)
			}
		}
	}

	return &UptimeRobotConfig{
		APIKey:             apiKey,
		Enabled:            getEnvBool("UPTIMEROBOT_ENABLED", true),
		Interval:           getEnvInt("UPTIMEROBOT_INTERVAL", 300), // Default 5 minutes
		AlertContacts:      alertContacts,
		AutoCreateMonitors: getEnvBool("UPTIMEROBOT_AUTO_CREATE", true),
	}
}

// Helper functions for environment variables
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue string) time.Duration {
	value := getEnvString(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	// Fallback to default
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func splitString(s, sep string) []string {
	return strings.Split(s, sep)
}

func trimString(s string) string {
	return strings.TrimSpace(s)
}
