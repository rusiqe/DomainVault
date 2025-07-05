package config

import (
	"os"
	"testing"
	"time"

	"github.com/rusiqe/domainvault/internal/types"
)

func TestLoad(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{"PORT", "DATABASE_URL", "SYNC_INTERVAL", "LOG_LEVEL", "GODADDY_API_KEY", "GODADDY_API_SECRET", "NAMECHEAP_API_KEY", "NAMECHEAP_USERNAME"}
	
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	
	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		validate func(*Config) error
	}{
		{
			name: "default configuration",
			envVars: map[string]string{},
			wantErr: false,
			validate: func(c *Config) error {
				if c.Port != 8080 {
					t.Errorf("Expected default port 8080, got %d", c.Port)
				}
				if c.LogLevel != "info" {
					t.Errorf("Expected default log level 'info', got %s", c.LogLevel)
				}
				if c.SyncInterval != time.Hour {
					t.Errorf("Expected default sync interval 1h, got %v", c.SyncInterval)
				}
				return nil
			},
		},
		{
			name: "custom port",
			envVars: map[string]string{
				"PORT": "9000",
			},
			wantErr: false,
			validate: func(c *Config) error {
				if c.Port != 9000 {
					t.Errorf("Expected port 9000, got %d", c.Port)
				}
				return nil
			},
		},
		{
			name: "invalid port - too high",
			envVars: map[string]string{
				"PORT": "99999",
			},
			wantErr: true,
			validate: nil,
		},
		{
			name: "invalid port - negative",
			envVars: map[string]string{
				"PORT": "-1",
			},
			wantErr: true,
			validate: nil,
		},
		{
			name: "custom sync interval",
			envVars: map[string]string{
				"SYNC_INTERVAL": "30m",
			},
			wantErr: false,
			validate: func(c *Config) error {
				if c.SyncInterval != 30*time.Minute {
					t.Errorf("Expected sync interval 30m, got %v", c.SyncInterval)
				}
				return nil
			},
		},
		{
			name: "sync interval too short",
			envVars: map[string]string{
				"SYNC_INTERVAL": "30s",
			},
			wantErr: true,
			validate: nil,
		},
		{
			name: "godaddy provider configured",
			envVars: map[string]string{
				"GODADDY_API_KEY":    "test-key",
				"GODADDY_API_SECRET": "test-secret",
			},
			wantErr: false,
			validate: func(c *Config) error {
				if len(c.Providers) == 0 {
					t.Error("Expected at least one provider")
					return nil
				}
				
				found := false
				for _, p := range c.Providers {
					if p.Name == "godaddy" && p.Enabled {
						found = true
						if p.Credentials["api_key"] != "test-key" {
							t.Errorf("Expected api_key 'test-key', got %v", p.Credentials["api_key"])
						}
					}
				}
				if !found {
					t.Error("Expected godaddy provider to be configured")
				}
				return nil
			},
		},
		{
			name: "namecheap provider configured",
			envVars: map[string]string{
				"NAMECHEAP_API_KEY": "test-key",
				"NAMECHEAP_USERNAME": "test-user",
			},
			wantErr: false,
			validate: func(c *Config) error {
				found := false
				for _, p := range c.Providers {
					if p.Name == "namecheap" && p.Enabled {
						found = true
						if p.Credentials["api_key"] != "test-key" {
							t.Errorf("Expected api_key 'test-key', got %v", p.Credentials["api_key"])
						}
						if p.Credentials["username"] != "test-user" {
							t.Errorf("Expected username 'test-user', got %v", p.Credentials["username"])
						}
					}
				}
				if !found {
					t.Error("Expected namecheap provider to be configured")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for this test
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			
			// Clean up env vars after test
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			config, err := Load()
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Load() unexpected error: %v", err)
				return
			}
			
			if tt.validate != nil {
				if err := tt.validate(config); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

func TestConfig_validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "valid config",
			config: Config{
				Port:         8080,
				DatabaseURL:  "postgres://user:pass@localhost/db",
				SyncInterval: time.Hour,
				LogLevel:     "info",
			},
			wantErr: nil,
		},
		{
			name: "invalid port - zero",
			config: Config{
				Port:         0,
				DatabaseURL:  "postgres://user:pass@localhost/db",
				SyncInterval: time.Hour,
			},
			wantErr: types.ErrInvalidConfig,
		},
		{
			name: "invalid port - too high",
			config: Config{
				Port:         70000,
				DatabaseURL:  "postgres://user:pass@localhost/db",
				SyncInterval: time.Hour,
			},
			wantErr: types.ErrInvalidConfig,
		},
		{
			name: "missing database URL",
			config: Config{
				Port:         8080,
				DatabaseURL:  "",
				SyncInterval: time.Hour,
			},
			wantErr: types.ErrMissingConfig,
		},
		{
			name: "sync interval too short",
			config: Config{
				Port:         8080,
				DatabaseURL:  "postgres://user:pass@localhost/db",
				SyncInterval: 30 * time.Second,
			},
			wantErr: types.ErrInvalidConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if err != tt.wantErr {
				t.Errorf("Config.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			name:         "valid integer",
			key:          "TEST_INT",
			defaultValue: 100,
			envValue:     "200",
			expected:     200,
		},
		{
			name:         "invalid integer - fallback to default",
			key:          "TEST_INT",
			defaultValue: 100,
			envValue:     "invalid",
			expected:     100,
		},
		{
			name:         "empty value - fallback to default",
			key:          "TEST_INT",
			defaultValue: 100,
			envValue:     "",
			expected:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			defer os.Unsetenv(tt.key)
			
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}
			
			result := getEnvInt(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvInt() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     time.Duration
	}{
		{
			name:         "valid duration",
			key:          "TEST_DURATION",
			defaultValue: "1h",
			envValue:     "30m",
			expected:     30 * time.Minute,
		},
		{
			name:         "invalid duration - fallback to default",
			key:          "TEST_DURATION",
			defaultValue: "1h",
			envValue:     "invalid",
			expected:     time.Hour,
		},
		{
			name:         "empty value - fallback to default",
			key:          "TEST_DURATION",
			defaultValue: "1h",
			envValue:     "",
			expected:     time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			defer os.Unsetenv(tt.key)
			
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}
			
			result := getEnvDuration(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvDuration() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
