package config

import (
	"os"
	"strconv"
	"time"

	"github.com/rusiqe/domainvault/internal/types"
)

// Config holds all application configuration
type Config struct {
	Port         int           `json:"port"`
	DatabaseURL  string        `json:"database_url"`
	SyncInterval time.Duration `json:"sync_interval"`
	LogLevel     string        `json:"log_level"`
	Providers    []ProviderConfig `json:"providers"`
}

// ProviderConfig holds configuration for a domain registrar
type ProviderConfig struct {
	Name        string                 `json:"name"`
	Enabled     bool                   `json:"enabled"`
	Credentials map[string]interface{} `json:"credentials"`
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Port:         getEnvInt("PORT", 8080),
		DatabaseURL:  getEnvString("DATABASE_URL", "postgres://user:password@localhost/domainvault?sslmode=disable"),
		SyncInterval: getEnvDuration("SYNC_INTERVAL", "1h"),
		LogLevel:     getEnvString("LOG_LEVEL", "info"),
		Providers:    loadProviders(),
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

	return providers
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
