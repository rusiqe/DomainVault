package dns

import (
	"fmt"
	"time"

	"github.com/rusiqe/domainvault/internal/types"
)

// DNSService handles DNS record operations
type DNSService struct {
	repo DNSRepository
}

// DNSRepository defines the interface for DNS data operations
type DNSRepository interface {
	CreateRecord(record *types.DNSRecord) error
	GetRecordsByDomain(domainID string) ([]types.DNSRecord, error)
	GetRecordByID(id string) (*types.DNSRecord, error)
	UpdateRecord(record *types.DNSRecord) error
	DeleteRecord(id string) error
	DeleteRecordsByDomain(domainID string) error
	BulkCreateRecords(records []types.DNSRecord) error
}

// NewDNSService creates a new DNS service
func NewDNSService(repo DNSRepository) *DNSService {
	return &DNSService{
		repo: repo,
	}
}

// CreateRecord creates a new DNS record
func (d *DNSService) CreateRecord(record *types.DNSRecord) error {
	if err := d.validateRecord(record); err != nil {
		return fmt.Errorf("invalid DNS record: %w", err)
	}

	now := time.Now()
	record.CreatedAt = now
	record.UpdatedAt = now

	return d.repo.CreateRecord(record)
}

// GetDomainRecords retrieves all DNS records for a domain
func (d *DNSService) GetDomainRecords(domainID string) ([]types.DNSRecord, error) {
	return d.repo.GetRecordsByDomain(domainID)
}

// GetRecord retrieves a specific DNS record
func (d *DNSService) GetRecord(id string) (*types.DNSRecord, error) {
	return d.repo.GetRecordByID(id)
}

// UpdateRecord updates an existing DNS record
func (d *DNSService) UpdateRecord(record *types.DNSRecord) error {
	if err := d.validateRecord(record); err != nil {
		return fmt.Errorf("invalid DNS record: %w", err)
	}

	record.UpdatedAt = time.Now()
	return d.repo.UpdateRecord(record)
}

// CreateOrUpdateRecord creates a new DNS record or updates existing one with same type and name
func (d *DNSService) CreateOrUpdateRecord(record types.DNSRecord) error {
	if err := d.validateRecord(&record); err != nil {
		return fmt.Errorf("invalid DNS record: %w", err)
	}

	// Check if a record with same type and name already exists
	existingRecords, err := d.repo.GetRecordsByDomain(record.DomainID)
	if err != nil {
		return fmt.Errorf("failed to get existing records: %w", err)
	}

	// Find existing record with same type and name
	for _, existing := range existingRecords {
		if existing.Type == record.Type && existing.Name == record.Name {
			// Update existing record
			record.ID = existing.ID
			record.CreatedAt = existing.CreatedAt
			record.UpdatedAt = time.Now()
			return d.repo.UpdateRecord(&record)
		}
	}

	// Create new record if no existing record found
	now := time.Now()
	record.CreatedAt = now
	record.UpdatedAt = now
	return d.repo.CreateRecord(&record)
}

// DeleteRecord deletes a DNS record
func (d *DNSService) DeleteRecord(id string) error {
	return d.repo.DeleteRecord(id)
}

// BulkUpdateRecords updates multiple DNS records for a domain
func (d *DNSService) BulkUpdateRecords(domainID string, records []types.DNSRecord) error {
	// Delete existing records for the domain
	if err := d.repo.DeleteRecordsByDomain(domainID); err != nil {
		return fmt.Errorf("failed to delete existing records: %w", err)
	}

	// Validate all records
	for i := range records {
		records[i].DomainID = domainID
		if err := d.validateRecord(&records[i]); err != nil {
			return fmt.Errorf("invalid DNS record at index %d: %w", i, err)
		}
		
		now := time.Now()
		records[i].CreatedAt = now
		records[i].UpdatedAt = now
	}

	// Create new records
	return d.repo.BulkCreateRecords(records)
}

// GetCommonRecordTemplates returns common DNS record templates
func (d *DNSService) GetCommonRecordTemplates() map[string][]types.DNSRecord {
	return map[string][]types.DNSRecord{
		"basic_website": {
			{Type: "A", Name: "@", Value: "192.168.1.1", TTL: 3600},
			{Type: "A", Name: "www", Value: "192.168.1.1", TTL: 3600},
			{Type: "CNAME", Name: "*", Value: "@", TTL: 3600},
		},
		"email_hosting": {
			{Type: "MX", Name: "@", Value: "mail.example.com", TTL: 3600, Priority: intPtr(10)},
			{Type: "TXT", Name: "@", Value: "v=spf1 include:_spf.example.com ~all", TTL: 3600},
			{Type: "CNAME", Name: "mail", Value: "mail.example.com", TTL: 3600},
		},
		"cdn_setup": {
			{Type: "CNAME", Name: "cdn", Value: "cdn.example.com", TTL: 3600},
			{Type: "CNAME", Name: "assets", Value: "cdn.example.com", TTL: 3600},
			{Type: "CNAME", Name: "static", Value: "cdn.example.com", TTL: 3600},
		},
	}
}

// validateRecord validates a DNS record
func (d *DNSService) validateRecord(record *types.DNSRecord) error {
	if record.DomainID == "" {
		return fmt.Errorf("domain ID is required")
	}
	
	if record.Type == "" {
		return fmt.Errorf("record type is required")
	}
	
	if record.Name == "" {
		return fmt.Errorf("record name is required")
	}
	
	if record.Value == "" {
		return fmt.Errorf("record value is required")
	}
	
	if record.TTL <= 0 {
		record.TTL = 3600 // Default TTL
	}

	// Validate specific record types
	switch record.Type {
	case "MX":
		if record.Priority == nil {
			return fmt.Errorf("MX records require priority")
		}
	case "SRV":
		if record.Priority == nil || record.Weight == nil || record.Port == nil {
			return fmt.Errorf("SRV records require priority, weight, and port")
		}
	case "A":
		// Could add IP validation here
	case "AAAA":
		// Could add IPv6 validation here
	case "CNAME", "TXT", "NS":
		// Basic validation is sufficient
	default:
		return fmt.Errorf("unsupported record type: %s", record.Type)
	}

	return nil
}

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}
