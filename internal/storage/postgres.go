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
		INSERT INTO domains (id, name, provider, expires_at, created_at, updated_at)
		VALUES (:id, :name, :provider, :expires_at, :created_at, :updated_at)
		ON CONFLICT (name) DO UPDATE SET
			provider = EXCLUDED.provider,
			expires_at = EXCLUDED.expires_at,
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
	query := "SELECT id, name, provider, expires_at, created_at, updated_at FROM domains ORDER BY created_at DESC"
	
	err := r.db.Select(&domains, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all domains: %w", err)
	}
	
	return domains, nil
}

// GetByID retrieves a domain by its ID
func (r *PostgresRepo) GetByID(id string) (*types.Domain, error) {
	var domain types.Domain
	query := "SELECT id, name, provider, expires_at, created_at, updated_at FROM domains WHERE id = $1"
	
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

	query := "SELECT id, name, provider, expires_at, created_at, updated_at FROM domains"

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
		SELECT id, name, provider, expires_at, created_at, updated_at 
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
