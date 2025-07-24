package types

import (
	"fmt"
	"os"
	"time"
)

// SecureProviderCredentials stores provider connection metadata with references to environment variables
// This improves security by not storing actual API keys in the database
type SecureProviderCredentials struct {
	ID                  string    `json:"id" db:"id"`
	Provider            string    `json:"provider" db:"provider"`                       // Provider type (godaddy, namecheap, hostinger)
	Name                string    `json:"name" db:"name"`                               // User-friendly name
	AccountName         string    `json:"account_name" db:"account_name"`               // Account identifier
	CredentialReference string    `json:"credential_reference" db:"credential_reference"` // Reference to env var (e.g., "GODADDY_DEFAULT", "NAMECHEAP_ACCOUNT1")
	Enabled             bool      `json:"enabled" db:"enabled"`
	ConnectionStatus    string    `json:"connection_status" db:"connection_status"`     // connected, error, testing
	LastSync            *time.Time `json:"last_sync,omitempty" db:"last_sync"`
	LastSyncError       *string   `json:"last_sync_error,omitempty" db:"last_sync_error"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// SecureProviderConnectionRequest represents a secure provider connection request
type SecureProviderConnectionRequest struct {
	Provider              string `json:"provider" binding:"required"`
	Name                  string `json:"name" binding:"required"`                    // User-friendly name
	AccountName           string `json:"account_name" binding:"required"`           // Account identifier
	CredentialReference   string `json:"credential_reference" binding:"required"`   // Reference to environment variables
	TestConnection        bool   `json:"test_connection"`                           // Test before saving
	AutoSync              bool   `json:"auto_sync"`                                 // Enable auto-sync
	SyncIntervalHours     int    `json:"sync_interval_hours"`                       // Auto-sync interval in hours
}

// PredefinedCredentialOption represents a predefined credential set available from environment
type PredefinedCredentialOption struct {
	Reference   string            `json:"reference"`    // Environment variable reference (e.g., "GODADDY_DEFAULT")
	DisplayName string            `json:"display_name"` // Human-readable name (e.g., "GoDaddy Production Account")
	Provider    string            `json:"provider"`     // Provider name
	Fields      map[string]string `json:"fields"`       // Field mappings to env vars
	Available   bool              `json:"available"`    // Whether all required env vars are set
}

// CredentialReferenceMap defines how credential references map to environment variables
var CredentialReferenceMap = map[string]PredefinedCredentialOption{
	// GoDaddy credential references
	"GODADDY_DEFAULT": {
		Reference:   "GODADDY_DEFAULT",
		DisplayName: "GoDaddy Production Account",
		Provider:    "godaddy",
		Fields: map[string]string{
			"api_key":    "GODADDY_API_KEY",
			"api_secret": "GODADDY_API_SECRET",
		},
	},
	"GODADDY_STAGING": {
		Reference:   "GODADDY_STAGING",
		DisplayName: "GoDaddy Staging Account",
		Provider:    "godaddy",
		Fields: map[string]string{
			"api_key":    "GODADDY_STAGING_API_KEY",
			"api_secret": "GODADDY_STAGING_API_SECRET",
		},
	},
	
	// Namecheap credential references
	"NAMECHEAP_DEFAULT": {
		Reference:   "NAMECHEAP_DEFAULT",
		DisplayName: "Namecheap Production Account",
		Provider:    "namecheap",
		Fields: map[string]string{
			"api_key":   "NAMECHEAP_API_KEY",
			"username":  "NAMECHEAP_USERNAME",
			"client_ip": "NAMECHEAP_CLIENT_IP",
		},
	},
	"NAMECHEAP_STAGING": {
		Reference:   "NAMECHEAP_STAGING",
		DisplayName: "Namecheap Staging Account",
		Provider:    "namecheap",
		Fields: map[string]string{
			"api_key":   "NAMECHEAP_STAGING_API_KEY",
			"username":  "NAMECHEAP_STAGING_USERNAME",
			"client_ip": "NAMECHEAP_STAGING_CLIENT_IP",
		},
	},
	
	// Hostinger credential references
	"HOSTINGER_DEFAULT": {
		Reference:   "HOSTINGER_DEFAULT",
		DisplayName: "Hostinger Production Account",
		Provider:    "hostinger",
		Fields: map[string]string{
			"api_key":   "HOSTINGER_API_KEY",
			"client_id": "HOSTINGER_CLIENT_ID",
		},
	},
	"HOSTINGER_STAGING": {
		Reference:   "HOSTINGER_STAGING",
		DisplayName: "Hostinger Staging Account",
		Provider:    "hostinger",
		Fields: map[string]string{
			"api_key":   "HOSTINGER_STAGING_API_KEY",
			"client_id": "HOSTINGER_STAGING_CLIENT_ID",
		},
	},
}

// GetCredentialOptions returns available credential options for a provider
func GetCredentialOptions(provider string) []PredefinedCredentialOption {
	var options []PredefinedCredentialOption
	
	for _, option := range CredentialReferenceMap {
		if option.Provider == provider {
			// Check if all required environment variables are available
			option.Available = true
			for _, envVar := range option.Fields {
				if GetenvSecure(envVar) == "" {
					option.Available = false
					break
				}
			}
			options = append(options, option)
		}
	}
	
	return options
}

// ResolveCredentials resolves a credential reference to actual values from environment variables
func ResolveCredentials(reference string) (map[string]string, error) {
	option, exists := CredentialReferenceMap[reference]
	if !exists {
		return nil, fmt.Errorf("credential reference %s not found", reference)
	}
	
	credentials := make(map[string]string)
	for field, envVar := range option.Fields {
		value := GetenvSecure(envVar)
		if value == "" && isRequiredField(option.Provider, field) {
			return nil, fmt.Errorf("required environment variable %s is not set", envVar)
		}
		if value != "" {
			credentials[field] = value
		}
	}
	
	return credentials, nil
}

// isRequiredField checks if a field is required for a provider
func isRequiredField(provider, field string) bool {
	requiredFields := map[string][]string{
		"godaddy":   {"api_key", "api_secret"},
		"namecheap": {"api_key", "username"},
		"hostinger": {"api_key"},
	}
	
	required, exists := requiredFields[provider]
	if !exists {
		return false
	}
	
	for _, reqField := range required {
		if reqField == field {
			return true
		}
	}
	return false
}

// GetenvSecure is a wrapper for os.Getenv that could be extended with decryption
// For now, it's a simple wrapper, but it provides a place for future security enhancements
func GetenvSecure(key string) string {
	// Future enhancement: Add decryption logic here
	// For now, just return the environment variable
	return getEnvVarValue(key)
}

// getEnvVarValue is an internal function to get environment variable values
// This can be mocked for testing
var getEnvVarValue = func(key string) string {
	return os.Getenv(key)
}
