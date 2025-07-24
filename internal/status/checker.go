package status

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rusiqe/domainvault/internal/types"
)

// StatusChecker handles HTTP status monitoring for domains
type StatusChecker struct {
	client  *http.Client
	timeout time.Duration
}

// NewStatusChecker creates a new status checker with default settings
func NewStatusChecker() *StatusChecker {
	return &StatusChecker{
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Don't follow redirects, treat them as status codes
				return http.ErrUseLastResponse
			},
		},
		timeout: 10 * time.Second,
	}
}

// CheckDomain checks the HTTP status of a domain
func (sc *StatusChecker) CheckDomain(domain *types.Domain) error {
	if domain == nil {
		return fmt.Errorf("domain is nil")
	}

	now := time.Now()
	url := fmt.Sprintf("http://%s", domain.Name)
	
	ctx, cancel := context.WithTimeout(context.Background(), sc.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		domain.HTTPStatus = nil
		domain.StatusMessage = stringPtr(fmt.Sprintf("Failed to create request: %v", err))
		domain.LastStatusCheck = &now
		return nil
	}

	// Set a reasonable user agent
	req.Header.Set("User-Agent", "DomainVault/1.0 Status Checker")

	resp, err := sc.client.Do(req)
	if err != nil {
		// Check if it's a timeout or connection error
		if ctx.Err() == context.DeadlineExceeded {
			domain.HTTPStatus = intPtr(408) // Request Timeout
			domain.StatusMessage = stringPtr("Request timeout")
		} else {
			domain.HTTPStatus = intPtr(0) // Connection failed
			domain.StatusMessage = stringPtr(fmt.Sprintf("Connection failed: %v", err))
		}
		domain.LastStatusCheck = &now
		return nil
	}
	defer resp.Body.Close()

	// Update domain with status information
	domain.HTTPStatus = &resp.StatusCode
	domain.StatusMessage = stringPtr(getStatusMessage(resp.StatusCode))
	domain.LastStatusCheck = &now

	return nil
}

// CheckDomains checks the HTTP status of multiple domains
func (sc *StatusChecker) CheckDomains(domains []types.Domain) error {
	for i := range domains {
		if err := sc.CheckDomain(&domains[i]); err != nil {
			return fmt.Errorf("failed to check domain %s: %w", domains[i].Name, err)
		}
		
		// Small delay between requests to be respectful
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// CheckDomainWithHTTPS also tries HTTPS if HTTP fails
func (sc *StatusChecker) CheckDomainWithHTTPS(domain *types.Domain) error {
	// First try HTTP
	if err := sc.CheckDomain(domain); err != nil {
		return err
	}

	// If HTTP failed (status 0 or 4xx/5xx), try HTTPS
	if domain.HTTPStatus != nil && (*domain.HTTPStatus == 0 || *domain.HTTPStatus >= 400) {
		httpsURL := fmt.Sprintf("https://%s", domain.Name)
		now := time.Now()
		
		ctx, cancel := context.WithTimeout(context.Background(), sc.timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", httpsURL, nil)
		if err != nil {
			return nil // Keep the HTTP result
		}

		req.Header.Set("User-Agent", "DomainVault/1.0 Status Checker")

		resp, err := sc.client.Do(req)
		if err != nil {
			return nil // Keep the HTTP result
		}
		defer resp.Body.Close()

		// Only update if HTTPS worked better than HTTP
		if resp.StatusCode < 400 {
			domain.HTTPStatus = &resp.StatusCode
			domain.StatusMessage = stringPtr(getStatusMessage(resp.StatusCode) + " (HTTPS)")
			domain.LastStatusCheck = &now
		}
	}

	return nil
}

// getStatusMessage returns a human-readable message for HTTP status codes
func getStatusMessage(statusCode int) string {
	switch {
	case statusCode == 0:
		return "Connection failed"
	case statusCode >= 200 && statusCode < 300:
		return "OK"
	case statusCode >= 300 && statusCode < 400:
		return "Redirected"
	case statusCode == 400:
		return "Bad request"
	case statusCode == 401:
		return "Unauthorized"
	case statusCode == 403:
		return "Access forbidden"
	case statusCode == 404:
		return "Domain not found"
	case statusCode == 408:
		return "Request timeout"
	case statusCode >= 400 && statusCode < 500:
		return "Client error"
	case statusCode == 500:
		return "Internal server error"
	case statusCode == 502:
		return "Bad gateway"
	case statusCode == 503:
		return "Service temporarily unavailable"
	case statusCode == 504:
		return "Gateway timeout"
	case statusCode >= 500:
		return "Server error"
	default:
		return fmt.Sprintf("HTTP %d", statusCode)
	}
}

// Helper functions for pointer creation
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

// StatusSummary provides aggregated status information
type StatusSummary struct {
	TotalChecked  int            `json:"total_checked"`
	StatusCounts  map[string]int `json:"status_counts"`
	LastCheckTime time.Time      `json:"last_check_time"`
}

// GetStatusSummary creates a summary from a list of domains
func GetStatusSummary(domains []types.Domain) *StatusSummary {
	summary := &StatusSummary{
		StatusCounts: make(map[string]int),
	}

	var lastCheck time.Time
	for _, domain := range domains {
		if domain.LastStatusCheck != nil {
			summary.TotalChecked++
			
			if domain.LastStatusCheck.After(lastCheck) {
				lastCheck = *domain.LastStatusCheck
			}

			if domain.HTTPStatus != nil {
				statusGroup := getStatusGroup(*domain.HTTPStatus)
				summary.StatusCounts[statusGroup]++
			} else {
				summary.StatusCounts["unchecked"]++
			}
		}
	}

	summary.LastCheckTime = lastCheck
	return summary
}

// getStatusGroup groups status codes into categories
func getStatusGroup(statusCode int) string {
	switch {
	case statusCode == 0:
		return "connection_failed"
	case statusCode >= 200 && statusCode < 300:
		return "success"
	case statusCode >= 300 && statusCode < 400:
		return "redirect"
	case statusCode >= 400 && statusCode < 500:
		return "client_error"
	case statusCode >= 500:
		return "server_error"
	default:
		return "unknown"
	}
}

// CheckWebsiteStatus checks the website status for multiple domains
func (sc *StatusChecker) CheckWebsiteStatus(request types.WebsiteStatusRequest) ([]types.WebsiteStatusResult, error) {
	results := make([]types.WebsiteStatusResult, 0, len(request.Domains))
	
	for _, domainName := range request.Domains {
		result := sc.checkSingleWebsiteStatus(domainName)
		results = append(results, result)
		
		// Small delay between requests to be respectful
		time.Sleep(100 * time.Millisecond)
	}
	
	return results, nil
}

// BulkCheckWebsiteStatus checks website status for multiple domains concurrently
func (sc *StatusChecker) BulkCheckWebsiteStatus(request types.WebsiteStatusRequest) ([]types.WebsiteStatusResult, error) {
	results := make([]types.WebsiteStatusResult, len(request.Domains))
	type result struct {
		index int
		status types.WebsiteStatusResult
	}
	
	resultChan := make(chan result, len(request.Domains))
	
	// Start goroutines for each domain check
	for i, domainName := range request.Domains {
		go func(index int, domain string) {
			status := sc.checkSingleWebsiteStatus(domain)
			resultChan <- result{index: index, status: status}
		}(i, domainName)
	}
	
	// Collect results
	for i := 0; i < len(request.Domains); i++ {
		result := <-resultChan
		results[result.index] = result.status
	}
	
	return results, nil
}

// checkSingleWebsiteStatus checks the status of a single website
func (sc *StatusChecker) checkSingleWebsiteStatus(domainName string) types.WebsiteStatusResult {
	now := time.Now()
	result := types.WebsiteStatusResult{
		Domain:      domainName,
		LastChecked: now,
	}
	
	// Try HTTP first
	httpResult := sc.performStatusCheck(fmt.Sprintf("http://%s", domainName))
	result.HTTPStatus = httpResult.statusCode
	result.StatusMessage = httpResult.message
	result.ResponseTime = httpResult.responseTime
	result.RedirectURL = httpResult.redirectURL
	result.Error = httpResult.error
	
	// If HTTP fails or returns an error, try HTTPS
	if httpResult.statusCode == 0 || httpResult.statusCode >= 400 {
		httpsResult := sc.performStatusCheck(fmt.Sprintf("https://%s", domainName))
		
		// Use HTTPS result if it's better
		if httpsResult.statusCode > 0 && httpsResult.statusCode < httpResult.statusCode {
			result.HTTPStatus = httpsResult.statusCode
			result.StatusMessage = httpsResult.message + " (HTTPS)"
			result.ResponseTime = httpsResult.responseTime
			result.RedirectURL = httpsResult.redirectURL
			result.Error = httpsResult.error
			result.SSLStatus = "valid"
		}
	} else if httpResult.statusCode >= 200 && httpResult.statusCode < 300 {
		// Also check HTTPS to see if SSL is available
		httpsResult := sc.performStatusCheck(fmt.Sprintf("https://%s", domainName))
		if httpsResult.statusCode >= 200 && httpsResult.statusCode < 300 {
			result.SSLStatus = "valid"
		} else {
			result.SSLStatus = "unavailable"
		}
	}
	
	return result
}

// statusCheckResult represents the result of a single status check
type statusCheckResult struct {
	statusCode   int
	message      string
	responseTime int64
	redirectURL  string
	error        string
}

// performStatusCheck performs the actual HTTP check
func (sc *StatusChecker) performStatusCheck(url string) statusCheckResult {
	start := time.Now()
	result := statusCheckResult{}
	
	ctx, cancel := context.WithTimeout(context.Background(), sc.timeout)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		result.error = fmt.Sprintf("Failed to create request: %v", err)
		result.message = "Request creation failed"
		return result
	}
	
	req.Header.Set("User-Agent", "DomainVault/1.0 Website Status Checker")
	
	resp, err := sc.client.Do(req)
	responseTime := time.Since(start).Milliseconds()
	result.responseTime = responseTime
	
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.statusCode = 408
			result.message = "Request timeout"
		} else {
			result.statusCode = 0
			result.message = "Connection failed"
			result.error = err.Error()
		}
		return result
	}
	defer resp.Body.Close()
	
	result.statusCode = resp.StatusCode
	result.message = getStatusMessage(resp.StatusCode)
	
	// Check for redirects
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		if location := resp.Header.Get("Location"); location != "" {
			result.redirectURL = location
		}
	}
	
	return result
}
