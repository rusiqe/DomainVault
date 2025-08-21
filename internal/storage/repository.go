package storage

import (
	"time"

	"github.com/rusiqe/domainvault/internal/types"
)

// DomainRepository defines the interface for domain data operations
type DomainRepository interface {
	// Core operations
	UpsertDomains(domains []types.Domain) error
	GetAll() ([]types.Domain, error)
	GetByID(id string) (*types.Domain, error)
	GetByFilter(filter types.DomainFilter) ([]types.Domain, error)
	GetDomainsByName(name string) ([]types.Domain, error)
	Delete(id string) error // Soft delete: sets visible=false
	Update(domain *types.Domain) error
	SetVisibility(id string, visible bool) error
	
	// Utility operations
	GetExpiring(threshold time.Duration) ([]types.Domain, error)
	GetSummary() (*types.DomainSummary, error)
	BulkRenew(domainIDs []string) error
	
	// User management
	CreateUser(user *types.User) error
	GetUserByUsername(username string) (*types.User, error)
	GetUserByID(id string) (*types.User, error)
	UpdateUser(user *types.User) error
	DeleteUser(id string) error
	UpdateLastLogin(userID string) error
	
	// Session management
	CreateSession(session *types.Session) error
	GetSessionByToken(token string) (*types.Session, error)
	DeleteSession(token string) error
	DeleteExpiredSessions() error
	
	// DNS management
	CreateRecord(record *types.DNSRecord) error
	GetRecordsByDomain(domainID string) ([]types.DNSRecord, error)
	GetRecordByID(id string) (*types.DNSRecord, error)
	UpdateRecord(record *types.DNSRecord) error
	DeleteRecord(id string) error
	DeleteRecordsByDomain(domainID string) error
	BulkCreateRecords(records []types.DNSRecord) error
	
	// Category management
	CreateCategory(category *types.Category) error
	GetAllCategories() ([]types.Category, error)
	GetCategoryByID(id string) (*types.Category, error)
	UpdateCategory(category *types.Category) error
	DeleteCategory(id string) error
	
	// Project management
	CreateProject(project *types.Project) error
	GetAllProjects() ([]types.Project, error)
	GetProjectByID(id string) (*types.Project, error)
	UpdateProject(project *types.Project) error
	DeleteProject(id string) error
	
	// Credentials management (legacy - to be deprecated)
	CreateCredentials(creds *types.ProviderCredentials) error
	GetAllCredentials() ([]types.ProviderCredentials, error)
	GetCredentialsByID(id string) (*types.ProviderCredentials, error)
	GetCredentialsByProvider(provider string) ([]types.ProviderCredentials, error)
	UpdateCredentials(creds *types.ProviderCredentials) error
	DeleteCredentials(id string) error
	
	// Secure credentials management (environment variable references)
	CreateSecureCredentials(creds *types.SecureProviderCredentials) error
	GetAllSecureCredentials() ([]types.SecureProviderCredentials, error)
	GetSecureCredentialsByID(id string) (*types.SecureProviderCredentials, error)
	GetSecureCredentialsByProvider(provider string) ([]types.SecureProviderCredentials, error)
	GetSecureCredentialsByReference(reference string) (*types.SecureProviderCredentials, error)
	UpdateSecureCredentials(creds *types.SecureProviderCredentials) error
	DeleteSecureCredentials(id string) error
	
	// Connection management
	Close() error
	Ping() error
}

// CategoryRepository defines the interface for category operations
type CategoryRepository interface {
	Create(category *types.Category) error
	GetAll() ([]types.Category, error)
	GetByID(id string) (*types.Category, error)
	Update(category *types.Category) error
	Delete(id string) error
}

// ProjectRepository defines the interface for project operations
type ProjectRepository interface {
	Create(project *types.Project) error
	GetAll() ([]types.Project, error)
	GetByID(id string) (*types.Project, error)
	Update(project *types.Project) error
	Delete(id string) error
}

// CredentialsRepository defines the interface for provider credentials operations
type CredentialsRepository interface {
	Create(creds *types.ProviderCredentials) error
	GetAll() ([]types.ProviderCredentials, error)
	GetByID(id string) (*types.ProviderCredentials, error)
	GetByProvider(provider string) ([]types.ProviderCredentials, error)
	Update(creds *types.ProviderCredentials) error
	Delete(id string) error
}
