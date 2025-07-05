package types

import (
	"testing"
	"time"
)

func TestDomain_Validate(t *testing.T) {
	tests := []struct {
		name    string
		domain  Domain
		wantErr error
	}{
		{
			name: "valid domain",
			domain: Domain{
				ID:       "test-id",
				Name:     "example.com",
				Provider: "godaddy",
			},
			wantErr: nil,
		},
		{
			name: "missing domain name",
			domain: Domain{
				ID:       "test-id",
				Name:     "",
				Provider: "godaddy",
			},
			wantErr: ErrInvalidDomainName,
		},
		{
			name: "missing provider",
			domain: Domain{
				ID:       "test-id",
				Name:     "example.com",
				Provider: "",
			},
			wantErr: ErrInvalidProvider,
		},
		{
			name: "empty domain",
			domain: Domain{},
			wantErr: ErrInvalidDomainName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.domain.Validate()
			if err != tt.wantErr {
				t.Errorf("Domain.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDomain_IsExpiringSoon(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name     string
		domain   Domain
		duration time.Duration
		want     bool
	}{
		{
			name: "expiring in 15 days, check 30 days",
			domain: Domain{
				ExpiresAt: now.AddDate(0, 0, 15),
			},
			duration: 30 * 24 * time.Hour,
			want:     true,
		},
		{
			name: "expiring in 45 days, check 30 days",
			domain: Domain{
				ExpiresAt: now.AddDate(0, 0, 45),
			},
			duration: 30 * 24 * time.Hour,
			want:     false,
		},
		{
			name: "already expired",
			domain: Domain{
				ExpiresAt: now.AddDate(0, 0, -1),
			},
			duration: 30 * 24 * time.Hour,
			want:     true,
		},
		{
			name: "expiring exactly at threshold",
			domain: Domain{
				ExpiresAt: now.Add(30 * 24 * time.Hour),
			},
			duration: 30 * 24 * time.Hour,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.domain.IsExpiringSoon(tt.duration)
			if got != tt.want {
				t.Errorf("Domain.IsExpiringSoon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDomain_DaysUntilExpiration(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name   string
		domain Domain
		want   int
	}{
		{
			name: "expires in 30 days",
			domain: Domain{
				ExpiresAt: now.AddDate(0, 0, 30),
			},
			want: 30,
		},
		{
			name: "expires in 1 day",
			domain: Domain{
				ExpiresAt: now.AddDate(0, 0, 1),
			},
			want: 1,
		},
		{
			name: "expired yesterday",
			domain: Domain{
				ExpiresAt: now.AddDate(0, 0, -1),
			},
			want: -1,
		},
		{
			name: "zero expiration date",
			domain: Domain{
				ExpiresAt: time.Time{},
			},
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.domain.DaysUntilExpiration()
			// Allow for small variations due to test execution time
			if abs(got-tt.want) > 1 {
				t.Errorf("Domain.DaysUntilExpiration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDomainFilter_Validation(t *testing.T) {
	// Test that DomainFilter struct can be created and used
	now := time.Now()
	after := now.AddDate(0, 0, -30)
	before := now.AddDate(0, 0, 30)

	filter := DomainFilter{
		Provider:      "godaddy",
		ExpiresAfter:  &after,
		ExpiresBefore: &before,
		Search:        "example",
		Limit:         10,
		Offset:        0,
	}

	if filter.Provider != "godaddy" {
		t.Errorf("Expected provider to be 'godaddy', got %s", filter.Provider)
	}
	
	if filter.Search != "example" {
		t.Errorf("Expected search to be 'example', got %s", filter.Search)
	}
	
	if filter.Limit != 10 {
		t.Errorf("Expected limit to be 10, got %d", filter.Limit)
	}
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
