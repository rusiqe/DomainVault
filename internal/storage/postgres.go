package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rusiqe/domainvault/internal/types"
)

// PostgresRepo implements DomainRepository for PostgreSQL
type PostgresRepo struct {
	db *sqlx.DB
}

// NewPostgresRepo creates a new PostgreSQL repository
func NewPostgresRepo(dsn string) (*PostgresRepo, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	repo := &PostgresRepo{db: db}
	
	// Test connection
	if err := repo.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return repo, nil
}

// Close closes the database connection
func (r *PostgresRepo) Close() error {
	return r.db.Close()
}

// Ping tests the database connection
func (r *PostgresRepo) Ping() error {
	return r.db.Ping()
}

// UpsertDomains inserts or updates multiple domains
func (r *PostgresRepo) UpsertDomains(domains []types.Domain) error {
	if len(domains) == 0 {
		return nil
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO domains (id, name, provider, expires_at, created_at, updated_at, category_id, project_id, auto_renew, renewal_price, status, tags, http_status, last_status_check, status_message)
		VALUES (:id, :name, :provider, :expires_at, :created_at, :updated_at, :category_id, :project_id, :auto_renew, :renewal_price, :status, :tags, :http_status, :last_status_check, :status_message)
		ON CONFLICT (name) DO UPDATE SET
			provider = EXCLUDED.provider,
			expires_at = EXCLUDED.expires_at,
			category_id = EXCLUDED.category_id,
			project_id = EXCLUDED.project_id,
			auto_renew = EXCLUDED.auto_renew,
			renewal_price = EXCLUDED.renewal_price,
			status = EXCLUDED.status,
			tags = EXCLUDED.tags,
			http_status = EXCLUDED.http_status,
			last_status_check = EXCLUDED.last_status_check,
			status_message = EXCLUDED.status_message,
			updated_at = NOW()
		RETURNING id`

	for i := range domains {
		// Generate UUID if not present
		if domains[i].ID == "" {
			domains[i].ID = uuid.New().String()
		}
		
		// Set timestamps
		now := time.Now()
		if domains[i].CreatedAt.IsZero() {
			domains[i].CreatedAt = now
		}
		domains[i].UpdatedAt = now

		_, err := tx.NamedExec(query, domains[i])
		if err != nil {
			return fmt.Errorf("failed to upsert domain %s: %w", domains[i].Name, err)
		}
	}

	return tx.Commit()
}

// GetAll retrieves all domains
func (r *PostgresRepo) GetAll() ([]types.Domain, error) {
	var domains []types.Domain
	query := "SELECT id, name, provider, expires_at, created_at, updated_at, category_id, project_id, auto_renew, renewal_price, status, tags, http_status, last_status_check, status_message FROM domains ORDER BY created_at DESC"
	
	err := r.db.Select(&domains, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all domains: %w", err)
	}
	
	return domains, nil
}

// GetByID retrieves a domain by its ID
func (r *PostgresRepo) GetByID(id string) (*types.Domain, error) {
	var domain types.Domain
	query := "SELECT id, name, provider, expires_at, created_at, updated_at, category_id, project_id, auto_renew, renewal_price, status, tags, http_status, last_status_check, status_message FROM domains WHERE id = $1"
	
	err := r.db.Get(&domain, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, types.ErrDomainNotFound
		}
		return nil, fmt.Errorf("failed to get domain by ID: %w", err)
	}
	
	return &domain, nil
}

// GetByFilter retrieves domains based on filter criteria
func (r *PostgresRepo) GetByFilter(filter types.DomainFilter) ([]types.Domain, error) {
	var domains []types.Domain
	var conditions []string
	var args []interface{}
	var argIndex int

	query := "SELECT id, name, provider, expires_at, created_at, updated_at, category_id, project_id, auto_renew, renewal_price, status, tags, http_status, last_status_check, status_message FROM domains"

	// Build WHERE conditions
	if filter.Provider != "" {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("provider = $%d", argIndex))
		args = append(args, filter.Provider)
	}

	if filter.ExpiresAfter != nil {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("expires_at > $%d", argIndex))
		args = append(args, *filter.ExpiresAfter)
	}

	if filter.ExpiresBefore != nil {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("expires_at < $%d", argIndex))
		args = append(args, *filter.ExpiresBefore)
	}

	if filter.Search != "" {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+filter.Search+"%")
	}

	if filter.CategoryID != nil {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *filter.CategoryID)
	}

	if filter.ProjectID != nil {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", argIndex))
		args = append(args, *filter.ProjectID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	// Add pagination
	if filter.Limit > 0 {
		argIndex++
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		argIndex++
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	err := r.db.Select(&domains, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get domains by filter: %w", err)
	}

	return domains, nil
}

// Delete removes a domain by ID
func (r *PostgresRepo) Delete(id string) error {
	result, err := r.db.Exec("DELETE FROM domains WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return types.ErrDomainNotFound
	}

	return nil
}

// GetExpiring retrieves domains expiring within the threshold
func (r *PostgresRepo) GetExpiring(threshold time.Duration) ([]types.Domain, error) {
	var domains []types.Domain
	query := `
		SELECT id, name, provider, expires_at, created_at, updated_at, category_id, project_id, auto_renew, renewal_price, status, tags, http_status, last_status_check, status_message
		FROM domains 
		WHERE expires_at <= NOW() + $1 
		ORDER BY expires_at ASC`

	err := r.db.Select(&domains, query, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to get expiring domains: %w", err)
	}

	return domains, nil
}

// GetSummary provides domain statistics
func (r *PostgresRepo) GetSummary() (*types.DomainSummary, error) {
	summary := &types.DomainSummary{
		ByProvider:  make(map[string]int),
		ExpiringIn:  make(map[string]int),
		LastSync:    time.Now(),
	}

	// Get total count
	err := r.db.Get(&summary.Total, "SELECT COUNT(*) FROM domains")
	if err != nil {
		return nil, fmt.Errorf("failed to get total domain count: %w", err)
	}

	// Get count by provider
	rows, err := r.db.Query("SELECT provider, COUNT(*) FROM domains GROUP BY provider")
	if err != nil {
		return nil, fmt.Errorf("failed to get domains by provider: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var provider string
		var count int
		if err := rows.Scan(&provider, &count); err != nil {
			return nil, err
		}
		summary.ByProvider[provider] = count
	}

	// Get expiring counts
	now := time.Now()
	expirationPeriods := map[string]time.Duration{
		"30_days":  30 * 24 * time.Hour,
		"90_days":  90 * 24 * time.Hour,
		"365_days": 365 * 24 * time.Hour,
	}

	for period, duration := range expirationPeriods {
		var count int
		query := "SELECT COUNT(*) FROM domains WHERE expires_at BETWEEN NOW() AND $1"
		err := r.db.Get(&count, query, now.Add(duration))
		if err != nil {
			return nil, fmt.Errorf("failed to get expiring count for %s: %w", period, err)
		}
		summary.ExpiringIn[period] = count
	}

	return summary, nil
}

// Update updates a single domain
func (r *PostgresRepo) Update(domain *types.Domain) error {
	domain.UpdatedAt = time.Now()
	query := `
		UPDATE domains 
		SET name = :name, provider = :provider, expires_at = :expires_at, 
		    category_id = :category_id, project_id = :project_id, auto_renew = :auto_renew, 
		    renewal_price = :renewal_price, status = :status, tags = :tags,
		    http_status = :http_status, last_status_check = :last_status_check, 
		    status_message = :status_message, updated_at = :updated_at
		WHERE id = :id`
	
	_, err := r.db.NamedExec(query, domain)
	if err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}
	
	return nil
}

// BulkRenew updates multiple domains (placeholder for future renewal logic)
func (r *PostgresRepo) BulkRenew(domainIDs []string) error {
	if len(domainIDs) == 0 {
		return nil
	}

	// For now, just update the updated_at timestamp
	// In the future, this would integrate with registrar APIs for actual renewal
	query := `UPDATE domains SET updated_at = NOW() WHERE id = ANY($1)`
	_, err := r.db.Exec(query, domainIDs)
	if err != nil {
		return fmt.Errorf("failed to bulk renew domains: %w", err)
	}

	return nil
}

// Category repository methods

// CreateCategory creates a new category
func (r *PostgresRepo) CreateCategory(category *types.Category) error {
	if category.ID == "" {
		category.ID = uuid.New().String()
	}
	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now
	
	query := `
		INSERT INTO categories (id, name, description, color, created_at, updated_at)
		VALUES (:id, :name, :description, :color, :created_at, :updated_at)`
	
	_, err := r.db.NamedExec(query, category)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	
	return nil
}

// GetAllCategories retrieves all categories
func (r *PostgresRepo) GetAllCategories() ([]types.Category, error) {
	var categories []types.Category
	query := "SELECT id, name, description, color, created_at, updated_at FROM categories ORDER BY name"
	
	err := r.db.Select(&categories, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all categories: %w", err)
	}
	
	return categories, nil
}

// GetCategoryByID retrieves a category by its ID
func (r *PostgresRepo) GetCategoryByID(id string) (*types.Category, error) {
	var category types.Category
	query := "SELECT id, name, description, color, created_at, updated_at FROM categories WHERE id = $1"
	
	err := r.db.Get(&category, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, types.ErrDomainNotFound // Reuse existing error
		}
		return nil, fmt.Errorf("failed to get category by ID: %w", err)
	}
	
	return &category, nil
}

// UpdateCategory updates a category
func (r *PostgresRepo) UpdateCategory(category *types.Category) error {
	category.UpdatedAt = time.Now()
	query := `
		UPDATE categories 
		SET name = :name, description = :description, color = :color, updated_at = :updated_at
		WHERE id = :id`
	
	_, err := r.db.NamedExec(query, category)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}
	
	return nil
}

// DeleteCategory deletes a category
func (r *PostgresRepo) DeleteCategory(id string) error {
	result, err := r.db.Exec("DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return types.ErrDomainNotFound
	}

	return nil
}

// Project repository methods

// CreateProject creates a new project
func (r *PostgresRepo) CreateProject(project *types.Project) error {
	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now
	
	query := `
		INSERT INTO projects (id, name, description, color, created_at, updated_at)
		VALUES (:id, :name, :description, :color, :created_at, :updated_at)`
	
	_, err := r.db.NamedExec(query, project)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	
	return nil
}

// GetAllProjects retrieves all projects
func (r *PostgresRepo) GetAllProjects() ([]types.Project, error) {
	var projects []types.Project
	query := "SELECT id, name, description, color, created_at, updated_at FROM projects ORDER BY name"
	
	err := r.db.Select(&projects, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all projects: %w", err)
	}
	
	return projects, nil
}

// GetProjectByID retrieves a project by its ID
func (r *PostgresRepo) GetProjectByID(id string) (*types.Project, error) {
	var project types.Project
	query := "SELECT id, name, description, color, created_at, updated_at FROM projects WHERE id = $1"
	
	err := r.db.Get(&project, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, types.ErrDomainNotFound
		}
		return nil, fmt.Errorf("failed to get project by ID: %w", err)
	}
	
	return &project, nil
}

// UpdateProject updates a project
func (r *PostgresRepo) UpdateProject(project *types.Project) error {
	project.UpdatedAt = time.Now()
	query := `
		UPDATE projects 
		SET name = :name, description = :description, color = :color, updated_at = :updated_at
		WHERE id = :id`
	
	_, err := r.db.NamedExec(query, project)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	
	return nil
}

// DeleteProject deletes a project
func (r *PostgresRepo) DeleteProject(id string) error {
	result, err := r.db.Exec("DELETE FROM projects WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return types.ErrDomainNotFound
	}

	return nil
}

// Credentials repository methods

// CreateCredentials creates new provider credentials
func (r *PostgresRepo) CreateCredentials(creds *types.ProviderCredentials) error {
	if creds.ID == "" {
		creds.ID = uuid.New().String()
	}
	now := time.Now()
	creds.CreatedAt = now
	creds.UpdatedAt = now
	
	query := `
		INSERT INTO provider_credentials (id, provider, name, account_name, credentials, enabled, connection_status, created_at, updated_at)
		VALUES (:id, :provider, :name, :account_name, :credentials, :enabled, :connection_status, :created_at, :updated_at)`
	
	_, err := r.db.NamedExec(query, creds)
	if err != nil {
		return fmt.Errorf("failed to create credentials: %w", err)
	}
	
	return nil
}

// GetAllCredentials retrieves all provider credentials
func (r *PostgresRepo) GetAllCredentials() ([]types.ProviderCredentials, error) {
	var credentials []types.ProviderCredentials
	query := `SELECT id, provider, name, account_name, credentials, enabled, connection_status, last_sync, last_sync_error, 
	          created_at, updated_at FROM provider_credentials ORDER BY provider, name`
	
	err := r.db.Select(&credentials, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all credentials: %w", err)
	}
	
	return credentials, nil
}

// GetCredentialsByID retrieves credentials by ID
func (r *PostgresRepo) GetCredentialsByID(id string) (*types.ProviderCredentials, error) {
	var creds types.ProviderCredentials
	query := `SELECT id, provider, name, account_name, credentials, enabled, connection_status, last_sync, last_sync_error, 
	          created_at, updated_at FROM provider_credentials WHERE id = $1`
	
	err := r.db.Get(&creds, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, types.ErrDomainNotFound
		}
		return nil, fmt.Errorf("failed to get credentials by ID: %w", err)
	}
	
	return &creds, nil
}

// GetCredentialsByProvider retrieves all credentials for a provider
func (r *PostgresRepo) GetCredentialsByProvider(provider string) ([]types.ProviderCredentials, error) {
	var credentials []types.ProviderCredentials
	query := `SELECT id, provider, name, account_name, credentials, enabled, connection_status, last_sync, last_sync_error, 
	          created_at, updated_at FROM provider_credentials WHERE provider = $1 ORDER BY name`
	
	err := r.db.Select(&credentials, query, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials by provider: %w", err)
	}
	
	return credentials, nil
}

// UpdateCredentials updates provider credentials
func (r *PostgresRepo) UpdateCredentials(creds *types.ProviderCredentials) error {
	creds.UpdatedAt = time.Now()
	query := `
		UPDATE provider_credentials 
		SET name = :name, account_name = :account_name, credentials = :credentials, enabled = :enabled, 
		    connection_status = :connection_status, last_sync = :last_sync, last_sync_error = :last_sync_error, updated_at = :updated_at
		WHERE id = :id`
	
	_, err := r.db.NamedExec(query, creds)
	if err != nil {
		return fmt.Errorf("failed to update credentials: %w", err)
	}
	
	return nil
}

// DeleteCredentials deletes provider credentials
func (r *PostgresRepo) DeleteCredentials(id string) error {
	result, err := r.db.Exec("DELETE FROM provider_credentials WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return types.ErrDomainNotFound
	}

	return nil
}

// User and session repository methods

// CreateUser creates a new user
func (r *PostgresRepo) CreateUser(user *types.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	
	query := `
		INSERT INTO users (id, username, email, password_hash, role, enabled, created_at, updated_at)
		VALUES (:id, :username, :email, :password_hash, :role, :enabled, :created_at, :updated_at)`
	
	_, err := r.db.NamedExec(query, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}

// GetUserByUsername retrieves a user by username
func (r *PostgresRepo) GetUserByUsername(username string) (*types.User, error) {
	var user types.User
	query := `SELECT id, username, email, password_hash, role, enabled, last_login, 
	          created_at, updated_at FROM users WHERE username = $1`
	
	err := r.db.Get(&user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, types.ErrDomainNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	
	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *PostgresRepo) GetUserByID(id string) (*types.User, error) {
	var user types.User
	query := `SELECT id, username, email, password_hash, role, enabled, last_login, 
	          created_at, updated_at FROM users WHERE id = $1`
	
	err := r.db.Get(&user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, types.ErrDomainNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	
	return &user, nil
}

// UpdateUser updates a user
func (r *PostgresRepo) UpdateUser(user *types.User) error {
	user.UpdatedAt = time.Now()
	query := `
		UPDATE users 
		SET username = :username, email = :email, password_hash = :password_hash, 
		    role = :role, enabled = :enabled, last_login = :last_login, updated_at = :updated_at
		WHERE id = :id`
	
	_, err := r.db.NamedExec(query, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	return nil
}

// DeleteUser deletes a user
func (r *PostgresRepo) DeleteUser(id string) error {
	result, err := r.db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return types.ErrDomainNotFound
	}

	return nil
}

// CreateSession creates a new session
func (r *PostgresRepo) CreateSession(session *types.Session) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	session.CreatedAt = time.Now()
	
	query := `
		INSERT INTO sessions (id, user_id, token, expires_at, created_at)
		VALUES (:id, :user_id, :token, :expires_at, :created_at)`
	
	_, err := r.db.NamedExec(query, session)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	
	return nil
}

// GetSessionByToken retrieves a session by token
func (r *PostgresRepo) GetSessionByToken(token string) (*types.Session, error) {
	var session types.Session
	query := "SELECT id, user_id, token, expires_at, created_at FROM sessions WHERE token = $1"
	
	err := r.db.Get(&session, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, types.ErrDomainNotFound
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}
	
	return &session, nil
}

// DeleteSession deletes a session
func (r *PostgresRepo) DeleteSession(token string) error {
	result, err := r.db.Exec("DELETE FROM sessions WHERE token = $1", token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return types.ErrDomainNotFound
	}

	return nil
}

// DeleteExpiredSessions removes expired sessions
func (r *PostgresRepo) DeleteExpiredSessions() error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE expires_at < NOW()")
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}

// UpdateLastLogin updates the last login time for a user
func (r *PostgresRepo) UpdateLastLogin(userID string) error {
	_, err := r.db.Exec("UPDATE users SET last_login = NOW() WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// DNS repository methods

// CreateRecord creates a new DNS record
func (r *PostgresRepo) CreateRecord(record *types.DNSRecord) error {
	if record.ID == "" {
		record.ID = uuid.New().String()
	}
	now := time.Now()
	record.CreatedAt = now
	record.UpdatedAt = now
	
	query := `
		INSERT INTO dns_records (id, domain_id, type, name, value, ttl, priority, weight, port, created_at, updated_at)
		VALUES (:id, :domain_id, :type, :name, :value, :ttl, :priority, :weight, :port, :created_at, :updated_at)`
	
	_, err := r.db.NamedExec(query, record)
	if err != nil {
		return fmt.Errorf("failed to create DNS record: %w", err)
	}
	
	return nil
}

// GetRecordsByDomain retrieves all DNS records for a domain
func (r *PostgresRepo) GetRecordsByDomain(domainID string) ([]types.DNSRecord, error) {
	var records []types.DNSRecord
	query := `SELECT id, domain_id, type, name, value, ttl, priority, weight, port, 
	          created_at, updated_at FROM dns_records WHERE domain_id = $1 ORDER BY type, name`
	
	err := r.db.Select(&records, query, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS records by domain: %w", err)
	}
	
	return records, nil
}

// GetRecordByID retrieves a DNS record by ID
func (r *PostgresRepo) GetRecordByID(id string) (*types.DNSRecord, error) {
	var record types.DNSRecord
	query := `SELECT id, domain_id, type, name, value, ttl, priority, weight, port, 
	          created_at, updated_at FROM dns_records WHERE id = $1`
	
	err := r.db.Get(&record, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, types.ErrDomainNotFound
		}
		return nil, fmt.Errorf("failed to get DNS record by ID: %w", err)
	}
	
	return &record, nil
}

// UpdateRecord updates a DNS record
func (r *PostgresRepo) UpdateRecord(record *types.DNSRecord) error {
	record.UpdatedAt = time.Now()
	query := `
		UPDATE dns_records 
		SET domain_id = :domain_id, type = :type, name = :name, value = :value, 
		    ttl = :ttl, priority = :priority, weight = :weight, port = :port, updated_at = :updated_at
		WHERE id = :id`
	
	_, err := r.db.NamedExec(query, record)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}
	
	return nil
}

// DeleteRecord deletes a DNS record
func (r *PostgresRepo) DeleteRecord(id string) error {
	result, err := r.db.Exec("DELETE FROM dns_records WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return types.ErrDomainNotFound
	}

	return nil
}

// DeleteRecordsByDomain deletes all DNS records for a domain
func (r *PostgresRepo) DeleteRecordsByDomain(domainID string) error {
	_, err := r.db.Exec("DELETE FROM dns_records WHERE domain_id = $1", domainID)
	if err != nil {
		return fmt.Errorf("failed to delete DNS records by domain: %w", err)
	}
	return nil
}

// BulkCreateRecords creates multiple DNS records
func (r *PostgresRepo) BulkCreateRecords(records []types.DNSRecord) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO dns_records (id, domain_id, type, name, value, ttl, priority, weight, port, created_at, updated_at)
		VALUES (:id, :domain_id, :type, :name, :value, :ttl, :priority, :weight, :port, :created_at, :updated_at)`

	for i := range records {
		// Generate UUID if not present
		if records[i].ID == "" {
			records[i].ID = uuid.New().String()
		}
		
		now := time.Now()
		records[i].CreatedAt = now
		records[i].UpdatedAt = now

		_, err := tx.NamedExec(query, records[i])
		if err != nil {
			return fmt.Errorf("failed to create DNS record %d: %w", i, err)
		}
	}

	return tx.Commit()
}
