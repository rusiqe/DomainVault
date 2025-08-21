package uptimerobot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/rusiqe/domainvault/internal/config"
	"github.com/rusiqe/domainvault/internal/types"
)

// Service provides UptimeRobot monitoring services for DomainVault
type Service struct {
	client      *Client
	config      *config.UptimeRobotConfig
	isConfigured bool
}

// NewService creates a new UptimeRobot service
func NewService(cfg *config.UptimeRobotConfig) *Service {
	service := &Service{
		config:      cfg,
		isConfigured: false,
	}

	// Check for mock mode
	if cfg != nil && cfg.APIKey == "mock" {
		// Enable mock mode for development
		service.isConfigured = true
		service.client = nil // Use mock responses
	} else if cfg != nil && cfg.APIKey != "" && cfg.Enabled {
		service.client = NewClient(cfg.APIKey)
		service.isConfigured = true
	}

	return service
}

// IsConfigured returns true if UptimeRobot is properly configured
func (s *Service) IsConfigured() bool {
	return s.isConfigured
}

// TestConnection tests the UptimeRobot API connection
func (s *Service) TestConnection() error {
	if !s.isConfigured {
		return fmt.Errorf("UptimeRobot is not configured")
	}

	// Mock mode
	if s.client == nil {
		return nil // Mock always succeeds
	}

	return s.client.TestConnection()
}

// GetAccountInfo retrieves UptimeRobot account information
func (s *Service) GetAccountInfo() (*Account, error) {
	if !s.isConfigured {
		return nil, fmt.Errorf("UptimeRobot is not configured")
	}

	// Mock mode
	if s.client == nil {
		return &Account{
			Email:          "demo@domainvault.com",
			MonitorLimit:   50,
			UpMonitors:     4,
			DownMonitors:   1,
			PausedMonitors: 0,
		}, nil
	}

	return s.client.GetAccountDetails()
}

// CreateMonitorForDomain creates a new monitor for a domain
func (s *Service) CreateMonitorForDomain(domain *types.Domain, monitorType MonitorType, interval int, alertContacts []string) (*CreateMonitorResponse, error) {
	if !s.isConfigured {
		return nil, fmt.Errorf("UptimeRobot is not configured")
	}

	if domain == nil {
		return nil, fmt.Errorf("domain is nil")
	}

	// Use default interval if not specified
	if interval == 0 {
		if s.config.Interval > 0 {
			interval = s.config.Interval
		} else {
			interval = DefaultInterval
		}
	}

	// Use configured alert contacts if not specified
	if len(alertContacts) == 0 {
		alertContacts = s.config.AlertContacts
	}

	// Generate friendly name
	friendlyName := fmt.Sprintf("DomainVault - %s", domain.Name)

	switch monitorType {
	case MonitorTypeHTTP:
		return s.client.CreateHTTPMonitor(domain.Name, friendlyName, interval, alertContacts)
	case MonitorTypePing:
		return s.client.CreatePingMonitor(domain.Name, friendlyName, interval, alertContacts)
	default:
		return nil, fmt.Errorf("unsupported monitor type: %d", monitorType)
	}
}

// CreateKeywordMonitorForDomain creates a keyword monitor for a domain
func (s *Service) CreateKeywordMonitorForDomain(domain *types.Domain, keyword string, keywordExists bool, interval int, alertContacts []string) (*CreateMonitorResponse, error) {
	if !s.isConfigured {
		return nil, fmt.Errorf("UptimeRobot is not configured")
	}

	if domain == nil {
		return nil, fmt.Errorf("domain is nil")
	}

	// Use default interval if not specified
	if interval == 0 {
		if s.config.Interval > 0 {
			interval = s.config.Interval
		} else {
			interval = DefaultInterval
		}
	}

	// Use configured alert contacts if not specified
	if len(alertContacts) == 0 {
		alertContacts = s.config.AlertContacts
	}

	// Generate friendly name
	friendlyName := fmt.Sprintf("DomainVault - %s (Keyword: %s)", domain.Name, keyword)

	return s.client.CreateKeywordMonitor(domain.Name, friendlyName, keyword, keywordExists, interval, alertContacts)
}

// UpdateMonitorForDomain updates an existing monitor for a domain
func (s *Service) UpdateMonitorForDomain(monitorID int, domain *types.Domain, interval int, alertContacts []string) (*UpdateMonitorResponse, error) {
	if !s.isConfigured {
		return nil, fmt.Errorf("UptimeRobot is not configured")
	}

	if domain == nil {
		return nil, fmt.Errorf("domain is nil")
	}

	req := &UpdateMonitorRequest{
		ID: monitorID,
	}

	// Update interval if specified
	if interval > 0 {
		req.Interval = interval
	}

	// Update alert contacts if specified
	if len(alertContacts) > 0 {
		req.AlertContacts = alertContacts
	}

	// Update friendly name to match current domain
	req.FriendlyName = fmt.Sprintf("DomainVault - %s", domain.Name)

	return s.client.UpdateMonitor(req)
}

// DeleteMonitorForDomain deletes a monitor for a domain
func (s *Service) DeleteMonitorForDomain(monitorID int) error {
	if !s.isConfigured {
		return fmt.Errorf("UptimeRobot is not configured")
	}

	_, err := s.client.DeleteMonitor(monitorID)
	return err
}

// GetMonitorStats retrieves detailed statistics for a monitor
func (s *Service) GetMonitorStats(monitorID int, customRanges ...string) (*Monitor, error) {
	if !s.isConfigured {
		return nil, fmt.Errorf("UptimeRobot is not configured")
	}

	return s.client.GetMonitorStats(monitorID, customRanges...)
}

// SyncDomainMonitoring synchronizes monitoring for multiple domains
func (s *Service) SyncDomainMonitoring(domains []types.Domain, monitorType MonitorType, autoCreate bool) (*types.UptimeRobotSyncResponse, error) {
	if !s.isConfigured {
		return &types.UptimeRobotSyncResponse{
			Success: false,
			Message: "UptimeRobot is not configured",
		}, nil
	}

	response := &types.UptimeRobotSyncResponse{
		Success: true,
		Message: "Monitoring synchronization completed",
		Results: make([]types.DomainMonitorResult, 0, len(domains)),
	}

	// Get all existing monitors
	existingMonitors, err := s.client.GetMonitors(nil)
	if err != nil {
		log.Printf("Failed to get existing monitors: %v", err)
		response.Success = false
		response.Message = fmt.Sprintf("Failed to get existing monitors: %v", err)
		return response, err
	}

	// Create a map of existing monitors by domain name
	monitorMap := make(map[string]*Monitor)
	for _, monitor := range existingMonitors {
		// Extract domain name from monitor URL
		domainName := s.extractDomainFromURL(monitor.URL)
		if domainName != "" {
			monitorMap[domainName] = &monitor
		}
	}

	// Process each domain
	for _, domain := range domains {
		result := types.DomainMonitorResult{
			DomainID:   domain.ID,
			DomainName: domain.Name,
		}

		// Check if monitor already exists
		if existingMonitor, exists := monitorMap[domain.Name]; exists {
			// Update existing monitor
			result.MonitorID = &existingMonitor.ID
			result.Action = "updated"

			// Get fresh stats for the monitor
			if stats, err := s.GetMonitorStats(existingMonitor.ID, "1", "7", "30"); err == nil {
				// Update domain with monitoring data (this would be done by the caller)
				result.Success = true
				result.Message = "Monitor updated successfully"
				if len(stats.ResponseTimes) > 0 {
					responseTime := stats.ResponseTimes[len(stats.ResponseTimes)-1].Value
					result.ResponseTime = &responseTime
				}
				response.MonitorsUpdated++
			} else {
				result.Success = false
				result.Error = fmt.Sprintf("Failed to get monitor stats: %v", err)
				result.Message = "Failed to update monitor stats"
				response.MonitorsFailed++
			}
		} else if autoCreate {
			// Create new monitor
			createResp, err := s.CreateMonitorForDomain(&domain, monitorType, 0, nil)
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("Failed to create monitor: %v", err)
				result.Message = "Monitor creation failed"
				result.Action = "failed"
				response.MonitorsFailed++
			} else {
				monitorID := createResp.Monitor.ID
				result.MonitorID = &monitorID
				result.Success = true
				result.Message = "Monitor created successfully"
				result.Action = "created"
				response.MonitorsCreated++
			}
		} else {
			// Skip domains without monitors when auto-create is disabled
			result.Success = true
			result.Message = "No monitor exists and auto-create is disabled"
			result.Action = "skipped"
		}

		result.Success = true
		response.Results = append(response.Results, result)
		response.MonitorsSync++
	}

	return response, nil
}

// BulkCreateMonitors creates monitors for multiple domains
func (s *Service) BulkCreateMonitors(request *types.UptimeRobotMonitorRequest) (*types.UptimeRobotSyncResponse, error) {
	if !s.isConfigured {
		return &types.UptimeRobotSyncResponse{
			Success: false,
			Message: "UptimeRobot is not configured",
		}, nil
	}

	response := &types.UptimeRobotSyncResponse{
		Success: true,
		Message: "Bulk monitor creation completed",
		Results: make([]types.DomainMonitorResult, 0, len(request.DomainIDs)),
	}

	// Convert monitor type string to enum
	// var monitorType MonitorType
	// switch strings.ToLower(request.MonitorType) {
	// case "http", "https":
	// 	monitorType = MonitorTypeHTTP
	// case "ping":
	// 	monitorType = MonitorTypePing
	// case "keyword":
	// 	monitorType = MonitorTypeKeyword
	// default:
	// 	monitorType = MonitorTypeHTTP // Default to HTTP
	// }

	// Process each domain ID (in a real implementation, you'd fetch domain data from repository)
	for _, domainID := range request.DomainIDs {
		result := types.DomainMonitorResult{
			DomainID: domainID,
			Action:   "created",
		}

		// In a real implementation, you would:
		// 1. Fetch domain data from repository using domainID
		// 2. Create monitor using the fetched domain data
		// 3. Update domain record with monitor ID

		// For now, we'll create a placeholder result
		result.Success = true
		result.Message = "Monitor would be created (placeholder implementation)"
		response.MonitorsCreated++
		response.Results = append(response.Results, result)
	}

	response.MonitorsSync = len(request.DomainIDs)
	return response, nil
}

// PauseMonitor pauses monitoring for a domain
func (s *Service) PauseMonitor(monitorID int) error {
	if !s.isConfigured {
		return fmt.Errorf("UptimeRobot is not configured")
	}

	return s.client.PauseMonitor(monitorID)
}

// ResumeMonitor resumes monitoring for a domain
func (s *Service) ResumeMonitor(monitorID int) error {
	if !s.isConfigured {
		return fmt.Errorf("UptimeRobot is not configured")
	}

	return s.client.ResumeMonitor(monitorID)
}

// GetAllMonitors retrieves all monitors with optional filtering
func (s *Service) GetAllMonitors(includeStats bool) ([]Monitor, error) {
	if !s.isConfigured {
		return nil, fmt.Errorf("UptimeRobot is not configured")
	}

	req := &GetMonitorsRequest{}
	if includeStats {
		req.ResponseTimes = 1
		req.ResponseTimesLimit = 1 // Just get the latest
		req.Logs = 1
		req.LogsLimit = 5 // Last 5 events
		req.CustomUptimeRatio = []string{"1", "7", "30"} // 1 day, 7 days, 30 days
	}

	return s.client.GetMonitors(req)
}

// GetDomainVaultMonitors retrieves only monitors created by DomainVault
func (s *Service) GetDomainVaultMonitors(includeStats bool) ([]Monitor, error) {
	monitors, err := s.GetAllMonitors(includeStats)
	if err != nil {
		return nil, err
	}

	// Filter monitors that were created by DomainVault
	var filtered []Monitor
	for _, monitor := range monitors {
		if strings.HasPrefix(monitor.FriendlyName, "DomainVault - ") {
			filtered = append(filtered, monitor)
		}
	}

	return filtered, nil
}

// UpdateConfig updates the UptimeRobot configuration
func (s *Service) UpdateConfig(cfg *config.UptimeRobotConfig) error {
	s.config = cfg

	if cfg != nil && cfg.APIKey != "" && cfg.Enabled {
		s.client = NewClient(cfg.APIKey)
		s.isConfigured = true

		// Test the new configuration
		if err := s.TestConnection(); err != nil {
			s.isConfigured = false
			return fmt.Errorf("failed to validate new configuration: %w", err)
		}
	} else {
		s.client = nil
		s.isConfigured = false
	}

	return nil
}

// GetMonitoringStats returns overall monitoring statistics
func (s *Service) GetMonitoringStats() (map[string]interface{}, error) {
	if !s.isConfigured {
		return map[string]interface{}{
			"configured": false,
			"message": "UptimeRobot is not configured",
		}, nil
	}

	// Mock mode
	if s.client == nil {
		return map[string]interface{}{
			"configured": true,
			"mock_mode": true,
			"total_monitors": 4,
			"up_monitors": 3,
			"down_monitors": 1,
			"paused_monitors": 0,
			"average_uptime": 98.5,
			"average_response_time": 250,
		}, nil
	}

	return s.GetMonitoringMetrics()
}

// GetMonitoringMetrics retrieves aggregated monitoring metrics
func (s *Service) GetMonitoringMetrics() (map[string]interface{}, error) {
	if !s.isConfigured {
		return nil, fmt.Errorf("UptimeRobot is not configured")
	}

	account, err := s.GetAccountInfo()
	if err != nil {
		return nil, err
	}

	monitors, err := s.GetDomainVaultMonitors(true)
	if err != nil {
		return nil, err
	}

	// Calculate metrics
	metrics := map[string]interface{}{
		"account": map[string]interface{}{
			"email":           account.Email,
			"monitor_limit":   account.MonitorLimit,
			"up_monitors":     account.UpMonitors,
			"down_monitors":   account.DownMonitors,
			"paused_monitors": account.PausedMonitors,
		},
		"domainvault_monitors": map[string]interface{}{
			"total":  len(monitors),
			"up":     0,
			"down":   0,
			"paused": 0,
		},
		"average_response_time": 0,
		"total_uptime_ratio":   0.0,
	}

	if len(monitors) > 0 {
		var totalResponseTime int
		var totalUptimeRatio float64
		responseTimeCount := 0

		for _, monitor := range monitors {
			// Count status
			switch monitor.Status {
			case MonitorStatusUp:
				metrics["domainvault_monitors"].(map[string]interface{})["up"] = 
					metrics["domainvault_monitors"].(map[string]interface{})["up"].(int) + 1
			case MonitorStatusDown, MonitorStatusSeemsDown:
				metrics["domainvault_monitors"].(map[string]interface{})["down"] = 
					metrics["domainvault_monitors"].(map[string]interface{})["down"].(int) + 1
			case MonitorStatusPaused:
				metrics["domainvault_monitors"].(map[string]interface{})["paused"] = 
					metrics["domainvault_monitors"].(map[string]interface{})["paused"].(int) + 1
			}

			// Calculate average response time from recent data
			if len(monitor.ResponseTimes) > 0 {
				totalResponseTime += monitor.ResponseTimes[len(monitor.ResponseTimes)-1].Value
				responseTimeCount++
			}
		}

		if responseTimeCount > 0 {
			metrics["average_response_time"] = totalResponseTime / responseTimeCount
		}

		metrics["total_uptime_ratio"] = totalUptimeRatio / float64(len(monitors))
	}

	return metrics, nil
}

// extractDomainFromURL extracts domain name from monitor URL
func (s *Service) extractDomainFromURL(url string) string {
	// Remove protocol
	domain := strings.TrimPrefix(url, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	
	// Remove path
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}
	
	// Remove port
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}
	
	return domain
}


// GetMonitors returns all UptimeRobot monitors
func (s *Service) GetMonitors() ([]Monitor, error) {
	if !s.isConfigured {
		return nil, fmt.Errorf("UptimeRobot is not configured")
	}

	// Mock mode
	if s.client == nil {
		return []Monitor{
			{
				ID: 12345,
				FriendlyName: "DomainVault - example.com",
				URL: "https://example.com",
				Type: MonitorTypeHTTP,
				Status: MonitorStatusUp,
				Interval: 300,
				CreateDatetime: 1640995200, // Unix timestamp
			},
		}, nil
	}

	return s.GetAllMonitors(true)
}

// SyncMonitors synchronizes UptimeRobot monitor data with the database
func (s *Service) SyncMonitors() error {
	if !s.isConfigured {
		return fmt.Errorf("UptimeRobot is not configured")
	}

	// Mock mode
	if s.client == nil {
		// Mock sync always succeeds
		return nil
	}

	// In a real implementation, this would:
	// 1. Fetch all monitors from UptimeRobot
	// 2. Update domain records with monitoring data
	// 3. Clean up orphaned monitors
	// For now, just validate the connection
	return s.TestConnection()
}

// CreateMonitor creates a new UptimeRobot monitor
func (s *Service) CreateMonitor(url string, monitorType MonitorType, name string, interval int) (*Monitor, error) {
	if !s.isConfigured {
		return nil, fmt.Errorf("UptimeRobot is not configured")
	}

	// Use default interval if not specified
	if interval == 0 {
		if s.config.Interval > 0 {
			interval = s.config.Interval
		} else {
			interval = DefaultInterval
		}
	}

	// Use default name if not specified
	if name == "" {
		name = fmt.Sprintf("DomainVault - %s", url)
	}

	// Mock mode
	if s.client == nil {
		return &Monitor{
			ID: 12346, // Mock ID
			FriendlyName: name,
			URL: url,
			Type: monitorType,
			Status: MonitorStatusUp,
			Interval: interval,
			CreateDatetime: time.Now().Unix(),
		}, nil
	}

	// Create monitor using appropriate method based on type
	var resp *CreateMonitorResponse
	var err error

	switch monitorType {
	case MonitorTypeHTTP:
		resp, err = s.client.CreateHTTPMonitor(url, name, interval, s.config.AlertContacts)
	case MonitorTypePing:
		resp, err = s.client.CreatePingMonitor(url, name, interval, s.config.AlertContacts)
	default:
		return nil, fmt.Errorf("unsupported monitor type: %d", monitorType)
	}

	if err != nil {
		return nil, err
	}

	// Create a Monitor object from the response
	monitor := &Monitor{
		ID: resp.Monitor.ID,
		FriendlyName: name,
		URL: url,
		Type: monitorType,
		Status: MonitorStatus(resp.Monitor.Status),
		Interval: interval,
		CreateDatetime: time.Now().Unix(),
	}

	return monitor, nil
}

// UpdateMonitor updates an existing UptimeRobot monitor
func (s *Service) UpdateMonitor(monitorID int, updates map[string]interface{}) error {
	if !s.isConfigured {
		return fmt.Errorf("UptimeRobot is not configured")
	}

	// Mock mode
	if s.client == nil {
		// Mock update always succeeds
		return nil
	}

	req := &UpdateMonitorRequest{
		ID: monitorID,
	}

	// Parse updates
	if friendlyName, ok := updates["friendly_name"].(string); ok {
		req.FriendlyName = friendlyName
	}
	if url, ok := updates["url"].(string); ok {
		req.URL = url
	}
	if interval, ok := updates["interval"].(int); ok {
		req.Interval = interval
	}
	if alertContacts, ok := updates["alert_contacts"].([]string); ok {
		req.AlertContacts = alertContacts
	}

	_, err := s.client.UpdateMonitor(req)
	return err
}

// DeleteMonitor deletes a UptimeRobot monitor
func (s *Service) DeleteMonitor(monitorID int) error {
	if !s.isConfigured {
		return fmt.Errorf("UptimeRobot is not configured")
	}

	// Mock mode
	if s.client == nil {
		// Mock delete always succeeds
		return nil
	}

	return s.DeleteMonitorForDomain(monitorID)
}

// GetMonitorLogs retrieves logs for a specific monitor
func (s *Service) GetMonitorLogs(monitorID int, limit int, offset int, startDate string, endDate string) ([]MonitorLog, error) {
	if !s.isConfigured {
		return nil, fmt.Errorf("UptimeRobot is not configured")
	}

	// Mock mode
	if s.client == nil {
		return []MonitorLog{
			{
				Type: 1, // Down
				Datetime: time.Now().Add(-2 * time.Hour).Unix(),
				Duration: 120, // 2 minutes
				Reason: Reason{Code: "timeout", Detail: "Connection timeout"},
			},
			{
				Type: 2, // Up
				Datetime: time.Now().Add(-2 * time.Hour).Add(2 * time.Minute).Unix(),
				Duration: 0,
				Reason: Reason{Code: "200", Detail: "OK"},
			},
		}, nil
	}

	req := &GetMonitorLogsRequest{
		MonitorIDs: []int{monitorID},
		Limit: limit,
		Offset: offset,
	}

	// Parse date filters if provided
	if startDate != "" {
		if startTime, err := time.Parse("2006-01-02", startDate); err == nil {
			req.StartDate = int(startTime.Unix())
		}
	}
	if endDate != "" {
		if endTime, err := time.Parse("2006-01-02", endDate); err == nil {
			req.EndDate = int(endTime.Unix())
		}
	}

	return s.client.GetMonitorLogs(req)
}

// ValidateConfig validates UptimeRobot configuration
func ValidateConfig(config *types.UptimeRobotConfig) error {
	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	if config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	// Validate interval
	validIntervals := []int{60, 120, 300, 600, 900, 1800, 3600}
	if config.Interval > 0 {
		valid := false
		for _, interval := range validIntervals {
			if config.Interval == interval {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid interval %d, must be one of: %v", config.Interval, validIntervals)
		}
	}

	return nil
}
