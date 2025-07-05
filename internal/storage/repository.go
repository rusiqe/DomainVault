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
	Delete(id string) error
	
	// Utility operations
	GetExpiring(threshold time.Duration) ([]types.Domain, error)
	GetSummary() (*types.DomainSummary, error)
	BulkRenew(domainIDs []string) error
	
	// Connection management
	Close() error
	Ping() error
}
