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
	Delete(id string) error
	Update(domain *types.Domain) error
	
	// Utility operations
	GetExpiring(threshold time.Duration) ([]types.Domain, error)
	GetSummary() (*types.DomainSummary, error)
	BulkRenew(domainIDs []string) error
	
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
