package storage

import (
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rusiqe/domainvault/internal/types"
)

// MockRepo implements DomainRepository for testing and development
type MockRepo struct {
	domains           map[string]types.Domain
	categories        map[string]types.Category
	projects          map[string]types.Project
	credentials       map[string]types.ProviderCredentials
	secureCredentials map[string]types.SecureProviderCredentials
	users             map[string]types.User
	sessions          map[string]types.Session
	dnsRecords        map[string]types.DNSRecord
	mu                sync.RWMutex
}

// NewMockRepo creates a new mock repository with sample data
func NewMockRepo() *MockRepo {
	repo := &MockRepo{
		domains:           make(map[string]types.Domain),
		categories:        make(map[string]types.Category),
		projects:          make(map[string]types.Project),
		credentials:       make(map[string]types.ProviderCredentials),
		secureCredentials: make(map[string]types.SecureProviderCredentials),
		users:             make(map[string]types.User),
		sessions:          make(map[string]types.Session),
		dnsRecords:        make(map[string]types.DNSRecord),
	}
	
	// Populate with sample data
	repo.populateSampleData()
	
	return repo
}

func (r *MockRepo) populateSampleData() {
	now := time.Now()
	
	// Sample categories
	categories := []types.Category{
		{ID: "cat1", Name: "Business", Description: "Business domains", Color: "#3498db", CreatedAt: now, UpdatedAt: now},
		{ID: "cat2", Name: "Personal", Description: "Personal projects", Color: "#e74c3c", CreatedAt: now, UpdatedAt: now},
		{ID: "cat3", Name: "E-commerce", Description: "Online stores", Color: "#2ecc71", CreatedAt: now, UpdatedAt: now},
	}
	for _, cat := range categories {
		r.categories[cat.ID] = cat
	}
	
	// Sample projects
	projects := []types.Project{
		{ID: "proj1", Name: "Main Website", Description: "Primary business website", Color: "#9b59b6", CreatedAt: now, UpdatedAt: now},
		{ID: "proj2", Name: "Side Projects", Description: "Personal side projects", Color: "#f39c12", CreatedAt: now, UpdatedAt: now},
	}
	for _, proj := range projects {
		r.projects[proj.ID] = proj
	}
	
	// Sample credentials
	creds := []types.ProviderCredentials{
		{
			ID:               "cred1",
			Provider:         "godaddy",
			Name:             "GoDaddy Main",
			AccountName:      "main@example.com",
			Credentials:      map[string]string{"api_key": "mock_key", "api_secret": "mock_secret"},
			Enabled:          true,
			ConnectionStatus: "connected",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "cred2",
			Provider:         "namecheap",
			Name:             "Namecheap Account",
			AccountName:      "user@example.com",
			Credentials:      map[string]string{"api_key": "mock_nc_key", "username": "mockuser"},
			Enabled:          true,
			ConnectionStatus: "connected",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "cred3",
			Provider:         "hostinger",
			Name:             "Hostinger Account",
			AccountName:      "support@example.com",
			Credentials:      map[string]string{"api_key": "mock_hostinger_key"},
			Enabled:          true,
			ConnectionStatus: "connected",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}
	for _, cred := range creds {
		r.credentials[cred.ID] = cred
	}
	
	// Sample domains
	domains := []types.Domain{
		{
			ID:          "dom1",
			Name:        "example.com",
			Provider:    "godaddy",
			ExpiresAt:   now.AddDate(1, 0, 0),
			CategoryID:  &[]string{"cat1"}[0],
			ProjectID:   &[]string{"proj1"}[0],
			AutoRenew:   true,
			RenewalPrice: &[]float64{12.99}[0],
			Status:      "active",
			Tags:        []string{"important", "business"},
			HTTPStatus:  &[]int{200}[0],
			LastStatusCheck: &now,
			StatusMessage:   &[]string{"OK"}[0],
			CreatedAt:   now.AddDate(0, -6, 0),
			UpdatedAt:   now,
		},
		{
			ID:          "dom2",
			Name:        "mystore.com",
			Provider:    "namecheap",
			ExpiresAt:   now.AddDate(0, 2, 0),
			CategoryID:  &[]string{"cat3"}[0],
			ProjectID:   &[]string{"proj1"}[0],
			AutoRenew:   false,
			RenewalPrice: &[]float64{15.88}[0],
			Status:      "active",
			Tags:        []string{"ecommerce", "urgent"},
			HTTPStatus:  &[]int{200}[0],
			LastStatusCheck: &now,
			StatusMessage:   &[]string{"OK"}[0],
			CreatedAt:   now.AddDate(0, -3, 0),
			UpdatedAt:   now,
		},
		{
			ID:          "dom3",
			Name:        "blog.example.org",
			Provider:    "godaddy",
			ExpiresAt:   now.AddDate(0, 0, 15),
			CategoryID:  &[]string{"cat2"}[0],
			ProjectID:   &[]string{"proj2"}[0],
			AutoRenew:   true,
			RenewalPrice: &[]float64{8.99}[0],
			Status:      "active",
			Tags:        []string{"blog", "personal"},
			HTTPStatus:  &[]int{404}[0],
			LastStatusCheck: &now,
			StatusMessage:   &[]string{"Not Found"}[0],
			CreatedAt:   now.AddDate(0, -1, 0),
			UpdatedAt:   now,
		},
		{
			ID:          "dom4",
			Name:        "testsite.net",
			Provider:    "namecheap",
			ExpiresAt:   now.AddDate(2, 0, 0),
			CategoryID:  &[]string{"cat2"}[0],
			AutoRenew:   false,
			RenewalPrice: &[]float64{22.00}[0],
			Status:      "active",
			Tags:        []string{"development", "testing"},
			HTTPStatus:  &[]int{500}[0],
			LastStatusCheck: &now,
			StatusMessage:   &[]string{"Internal Server Error"}[0],
			CreatedAt:   now.AddDate(0, -8, 0),
			UpdatedAt:   now,
		},
		{
			ID:          "dom5",
			Name:        "portfolio.dev",
			Provider:    "godaddy",
			ExpiresAt:   now.AddDate(0, 1, 0),
			CategoryID:  &[]string{"cat2"}[0],
			ProjectID:   &[]string{"proj2"}[0],
			AutoRenew:   true,
			RenewalPrice: &[]float64{35.99}[0],
			Status:      "active",
			Tags:        []string{"portfolio", "professional"},
			HTTPStatus:  &[]int{200}[0],
			LastStatusCheck: &now,
			StatusMessage:   &[]string{"OK"}[0],
			CreatedAt:   now.AddDate(0, -4, 0),
			UpdatedAt:   now,
		},
		{
			ID:          "dom6",
			Name:        "hosting-demo.com",
			Provider:    "hostinger",
			ExpiresAt:   now.AddDate(0, 8, 0),
			CategoryID:  &[]string{"cat1"}[0],
			ProjectID:   &[]string{"proj1"}[0],
			AutoRenew:   true,
			RenewalPrice: &[]float64{9.99}[0],
			Status:      "active",
			Tags:        []string{"hosting", "demo"},
			HTTPStatus:  &[]int{200}[0],
			LastStatusCheck: &now,
			StatusMessage:   &[]string{"OK"}[0],
			CreatedAt:   now.AddDate(0, -2, 0),
			UpdatedAt:   now,
		},
	}
	for _, domain := range domains {
		r.domains[domain.ID] = domain
	}
	
	// Sample admin user
	user := types.User{
		ID:           "user1",
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: "$2a$10$fWnAAUOrEFaEh.yJuMqb7.wJS6QR5p0pAl0f23Fn8cP7.5EwudlEa", // "admin123"
		Role:         "admin",
		Enabled:      true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	r.users[user.ID] = user
	
	// Sample DNS records
	dnsRecords := []types.DNSRecord{
		{
			ID:        "dns1",
			DomainID:  "dom1",
			Type:      "A",
			Name:      "@",
			Value:     "192.0.2.1",
			TTL:       3600,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "dns2",
			DomainID:  "dom1",
			Type:      "A",
			Name:      "www",
			Value:     "192.0.2.1",
			TTL:       3600,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "dns3",
			DomainID:  "dom1",
			Type:      "MX",
			Name:      "@",
			Value:     "mail.example.com",
			TTL:       3600,
			Priority:  &[]int{10}[0],
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "dns4",
			DomainID:  "dom2",
			Type:      "A",
			Name:      "@",
			Value:     "192.0.2.2",
			TTL:       1800,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "dns5",
			DomainID:  "dom2",
			Type:      "CNAME",
			Name:      "shop",
			Value:     "shopify.com",
			TTL:       3600,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	for _, record := range dnsRecords {
		r.dnsRecords[record.ID] = record
	}
}

// Domain repository methods
func (r *MockRepo) UpsertDomains(domains []types.Domain) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for _, domain := range domains {
		if domain.ID == "" {
			domain.ID = uuid.New().String()
		}
		if domain.CreatedAt.IsZero() {
			domain.CreatedAt = time.Now()
		}
		domain.UpdatedAt = time.Now()
		r.domains[domain.ID] = domain
	}
	return nil
}

func (r *MockRepo) GetAll() ([]types.Domain, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	domains := make([]types.Domain, 0, len(r.domains))
	for _, domain := range r.domains {
		domains = append(domains, domain)
	}
	return domains, nil
}

func (r *MockRepo) GetByID(id string) (*types.Domain, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	domain, exists := r.domains[id]
	if !exists {
		return nil, types.ErrDomainNotFound
	}
	return &domain, nil
}

func (r *MockRepo) GetByFilter(filter types.DomainFilter) ([]types.Domain, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var domains []types.Domain
	for _, domain := range r.domains {
		if r.matchesFilter(domain, filter) {
			domains = append(domains, domain)
		}
	}
	
	// Apply limit and offset
	if filter.Offset > 0 && filter.Offset < len(domains) {
		domains = domains[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(domains) {
		domains = domains[:filter.Limit]
	}
	
	return domains, nil
}

func (r *MockRepo) matchesFilter(domain types.Domain, filter types.DomainFilter) bool {
	if filter.Provider != "" && domain.Provider != filter.Provider {
		return false
	}
	if filter.ExpiresAfter != nil && domain.ExpiresAt.Before(*filter.ExpiresAfter) {
		return false
	}
	if filter.ExpiresBefore != nil && domain.ExpiresAt.After(*filter.ExpiresBefore) {
		return false
	}
	if filter.Search != "" && !strings.Contains(strings.ToLower(domain.Name), strings.ToLower(filter.Search)) {
		return false
	}
	if filter.CategoryID != nil && (domain.CategoryID == nil || *domain.CategoryID != *filter.CategoryID) {
		return false
	}
	if filter.ProjectID != nil && (domain.ProjectID == nil || *domain.ProjectID != *filter.ProjectID) {
		return false
	}
	return true
}

func (r *MockRepo) GetDomainsByName(name string) ([]types.Domain, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var domains []types.Domain
	for _, domain := range r.domains {
		if domain.Name == name {
			domains = append(domains, domain)
		}
	}
	return domains, nil
}

func (r *MockRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.domains[id]; !exists {
		return types.ErrDomainNotFound
	}
	delete(r.domains, id)
	return nil
}

func (r *MockRepo) Update(domain *types.Domain) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.domains[domain.ID]; !exists {
		return types.ErrDomainNotFound
	}
	domain.UpdatedAt = time.Now()
	r.domains[domain.ID] = *domain
	return nil
}

func (r *MockRepo) GetExpiring(threshold time.Duration) ([]types.Domain, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	cutoff := time.Now().Add(threshold)
	var domains []types.Domain
	for _, domain := range r.domains {
		if domain.ExpiresAt.Before(cutoff) {
			domains = append(domains, domain)
		}
	}
	return domains, nil
}

func (r *MockRepo) GetSummary() (*types.DomainSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	summary := &types.DomainSummary{
		Total:      len(r.domains),
		ByProvider: make(map[string]int),
		ExpiringIn: make(map[string]int),
		LastSync:   time.Now(),
	}
	
	now := time.Now()
	for _, domain := range r.domains {
		summary.ByProvider[domain.Provider]++
		
		if domain.ExpiresAt.Before(now.AddDate(0, 0, 30)) {
			summary.ExpiringIn["30_days"]++
		}
		if domain.ExpiresAt.Before(now.AddDate(0, 0, 90)) {
			summary.ExpiringIn["90_days"]++
		}
		if domain.ExpiresAt.Before(now.AddDate(1, 0, 0)) {
			summary.ExpiringIn["365_days"]++
		}
	}
	
	return summary, nil
}

func (r *MockRepo) BulkRenew(domainIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for _, id := range domainIDs {
		if domain, exists := r.domains[id]; exists {
			domain.UpdatedAt = time.Now()
			r.domains[id] = domain
		}
	}
	return nil
}

func (r *MockRepo) Close() error {
	return nil
}

func (r *MockRepo) Ping() error {
	return nil
}

// Category repository methods
func (r *MockRepo) CreateCategory(category *types.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if category.ID == "" {
		category.ID = uuid.New().String()
	}
	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now
	r.categories[category.ID] = *category
	return nil
}

func (r *MockRepo) GetAllCategories() ([]types.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	categories := make([]types.Category, 0, len(r.categories))
	for _, category := range r.categories {
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *MockRepo) GetCategoryByID(id string) (*types.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	category, exists := r.categories[id]
	if !exists {
		return nil, types.ErrDomainNotFound
	}
	return &category, nil
}

func (r *MockRepo) UpdateCategory(category *types.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.categories[category.ID]; !exists {
		return types.ErrDomainNotFound
	}
	category.UpdatedAt = time.Now()
	r.categories[category.ID] = *category
	return nil
}

func (r *MockRepo) DeleteCategory(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.categories[id]; !exists {
		return types.ErrDomainNotFound
	}
	delete(r.categories, id)
	return nil
}

// Project repository methods
func (r *MockRepo) CreateProject(project *types.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now
	r.projects[project.ID] = *project
	return nil
}

func (r *MockRepo) GetAllProjects() ([]types.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	projects := make([]types.Project, 0, len(r.projects))
	for _, project := range r.projects {
		projects = append(projects, project)
	}
	return projects, nil
}

func (r *MockRepo) GetProjectByID(id string) (*types.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	project, exists := r.projects[id]
	if !exists {
		return nil, types.ErrDomainNotFound
	}
	return &project, nil
}

func (r *MockRepo) UpdateProject(project *types.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.projects[project.ID]; !exists {
		return types.ErrDomainNotFound
	}
	project.UpdatedAt = time.Now()
	r.projects[project.ID] = *project
	return nil
}

func (r *MockRepo) DeleteProject(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.projects[id]; !exists {
		return types.ErrDomainNotFound
	}
	delete(r.projects, id)
	return nil
}

// Credentials repository methods
func (r *MockRepo) CreateCredentials(creds *types.ProviderCredentials) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if creds.ID == "" {
		creds.ID = uuid.New().String()
	}
	now := time.Now()
	creds.CreatedAt = now
	creds.UpdatedAt = now
	r.credentials[creds.ID] = *creds
	return nil
}

func (r *MockRepo) GetAllCredentials() ([]types.ProviderCredentials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	credentials := make([]types.ProviderCredentials, 0, len(r.credentials))
	for _, cred := range r.credentials {
		credentials = append(credentials, cred)
	}
	return credentials, nil
}

func (r *MockRepo) GetCredentialsByID(id string) (*types.ProviderCredentials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	creds, exists := r.credentials[id]
	if !exists {
		return nil, types.ErrDomainNotFound
	}
	return &creds, nil
}

func (r *MockRepo) GetCredentialsByProvider(provider string) ([]types.ProviderCredentials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var credentials []types.ProviderCredentials
	for _, cred := range r.credentials {
		if cred.Provider == provider {
			credentials = append(credentials, cred)
		}
	}
	return credentials, nil
}

func (r *MockRepo) UpdateCredentials(creds *types.ProviderCredentials) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.credentials[creds.ID]; !exists {
		return types.ErrDomainNotFound
	}
	creds.UpdatedAt = time.Now()
	r.credentials[creds.ID] = *creds
	return nil
}

func (r *MockRepo) DeleteCredentials(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.credentials[id]; !exists {
		return types.ErrDomainNotFound
	}
	delete(r.credentials, id)
	return nil
}

// User repository methods
func (r *MockRepo) CreateUser(user *types.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	r.users[user.ID] = *user
	return nil
}

func (r *MockRepo) GetUserByUsername(username string) (*types.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, user := range r.users {
		if user.Username == username {
			return &user, nil
		}
	}
	return nil, types.ErrDomainNotFound
}

func (r *MockRepo) GetUserByID(id string) (*types.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	user, exists := r.users[id]
	if !exists {
		return nil, types.ErrDomainNotFound
	}
	return &user, nil
}

func (r *MockRepo) UpdateUser(user *types.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.users[user.ID]; !exists {
		return types.ErrDomainNotFound
	}
	user.UpdatedAt = time.Now()
	r.users[user.ID] = *user
	return nil
}

func (r *MockRepo) DeleteUser(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.users[id]; !exists {
		return types.ErrDomainNotFound
	}
	delete(r.users, id)
	return nil
}

func (r *MockRepo) UpdateLastLogin(userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if user, exists := r.users[userID]; exists {
		now := time.Now()
		user.LastLogin = &now
		r.users[userID] = user
	}
	return nil
}

// Session repository methods
func (r *MockRepo) CreateSession(session *types.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	session.CreatedAt = time.Now()
	r.sessions[session.ID] = *session
	return nil
}

func (r *MockRepo) GetSessionByToken(token string) (*types.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, session := range r.sessions {
		if session.Token == token {
			return &session, nil
		}
	}
	return nil, types.ErrDomainNotFound
}

func (r *MockRepo) DeleteSession(token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for id, session := range r.sessions {
		if session.Token == token {
			delete(r.sessions, id)
			return nil
		}
	}
	return types.ErrDomainNotFound
}

func (r *MockRepo) DeleteExpiredSessions() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	now := time.Now()
	for id, session := range r.sessions {
		if session.ExpiresAt.Before(now) {
			delete(r.sessions, id)
		}
	}
	return nil
}

// DNS repository methods
func (r *MockRepo) CreateRecord(record *types.DNSRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if record.ID == "" {
		record.ID = uuid.New().String()
	}
	now := time.Now()
	record.CreatedAt = now
	record.UpdatedAt = now
	r.dnsRecords[record.ID] = *record
	return nil
}

func (r *MockRepo) GetRecordsByDomain(domainID string) ([]types.DNSRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var records []types.DNSRecord
	for _, record := range r.dnsRecords {
		if record.DomainID == domainID {
			records = append(records, record)
		}
	}
	return records, nil
}

func (r *MockRepo) GetRecordByID(id string) (*types.DNSRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	record, exists := r.dnsRecords[id]
	if !exists {
		return nil, types.ErrDomainNotFound
	}
	return &record, nil
}

func (r *MockRepo) UpdateRecord(record *types.DNSRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.dnsRecords[record.ID]; !exists {
		return types.ErrDomainNotFound
	}
	record.UpdatedAt = time.Now()
	r.dnsRecords[record.ID] = *record
	return nil
}

func (r *MockRepo) DeleteRecord(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.dnsRecords[id]; !exists {
		return types.ErrDomainNotFound
	}
	delete(r.dnsRecords, id)
	return nil
}

func (r *MockRepo) DeleteRecordsByDomain(domainID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for id, record := range r.dnsRecords {
		if record.DomainID == domainID {
			delete(r.dnsRecords, id)
		}
	}
	return nil
}

func (r *MockRepo) BulkCreateRecords(records []types.DNSRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for _, record := range records {
		if record.ID == "" {
			record.ID = uuid.New().String()
		}
		now := time.Now()
		record.CreatedAt = now
		record.UpdatedAt = now
		r.dnsRecords[record.ID] = record
	}
	return nil
}

// Secure credentials management methods
func (r *MockRepo) CreateSecureCredentials(creds *types.SecureProviderCredentials) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if creds.ID == "" {
		creds.ID = uuid.New().String()
	}
	now := time.Now()
	creds.CreatedAt = now
	creds.UpdatedAt = now
	r.secureCredentials[creds.ID] = *creds
	return nil
}

func (r *MockRepo) GetAllSecureCredentials() ([]types.SecureProviderCredentials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var creds []types.SecureProviderCredentials
	for _, cred := range r.secureCredentials {
		creds = append(creds, cred)
	}
	return creds, nil
}

func (r *MockRepo) GetSecureCredentialsByID(id string) (*types.SecureProviderCredentials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	cred, exists := r.secureCredentials[id]
	if !exists {
		return nil, types.ErrDomainNotFound
	}
	return &cred, nil
}

func (r *MockRepo) GetSecureCredentialsByProvider(provider string) ([]types.SecureProviderCredentials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var creds []types.SecureProviderCredentials
	for _, cred := range r.secureCredentials {
		if cred.Provider == provider {
			creds = append(creds, cred)
		}
	}
	return creds, nil
}

func (r *MockRepo) GetSecureCredentialsByReference(reference string) (*types.SecureProviderCredentials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, cred := range r.secureCredentials {
		if cred.CredentialReference == reference {
			return &cred, nil
		}
	}
	return nil, types.ErrDomainNotFound
}

func (r *MockRepo) UpdateSecureCredentials(creds *types.SecureProviderCredentials) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.secureCredentials[creds.ID]; !exists {
		return types.ErrDomainNotFound
	}
	creds.UpdatedAt = time.Now()
	r.secureCredentials[creds.ID] = *creds
	return nil
}

func (r *MockRepo) DeleteSecureCredentials(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.secureCredentials[id]; !exists {
		return types.ErrDomainNotFound
	}
	delete(r.secureCredentials, id)
	return nil
}

// NewRepo creates a repository instance based on the connection string
func NewRepo(dsn string) (DomainRepository, error) {
	if strings.HasPrefix(dsn, "mock://") {
		return NewMockRepo(), nil
	}
	return NewPostgresRepo(dsn)
}
