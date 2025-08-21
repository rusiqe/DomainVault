package uptimerobot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// BaseURL is the UptimeRobot API base URL
	BaseURL = "https://api.uptimerobot.com/v2"
	
	// Default timeout for HTTP requests
	DefaultTimeout = 30 * time.Second
	
	// Default interval for monitors (5 minutes)
	DefaultInterval = 300
)

// Client represents an UptimeRobot API client
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new UptimeRobot API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL: BaseURL,
	}
}

// NewClientWithHTTPClient creates a new UptimeRobot API client with custom HTTP client
func NewClientWithHTTPClient(apiKey string, httpClient *http.Client) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: httpClient,
		baseURL:    BaseURL,
	}
}

// SetBaseURL sets a custom base URL for the API (useful for testing)
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// makeRequest makes an HTTP POST request to the UptimeRobot API
func (c *Client) makeRequest(endpoint string, params map[string]interface{}) ([]byte, error) {
	if params == nil {
		params = make(map[string]interface{})
	}
	
	// Add API key to all requests
	params["api_key"] = c.apiKey
	params["format"] = "json"
	
	// Convert params to form data
	data := url.Values{}
	for key, value := range params {
		switch v := value.(type) {
		case string:
			data.Set(key, v)
		case int:
			data.Set(key, strconv.Itoa(v))
		case []string:
			data.Set(key, strings.Join(v, "-"))
		case []int:
			intStrings := make([]string, len(v))
			for i, num := range v {
				intStrings[i] = strconv.Itoa(num)
			}
			data.Set(key, strings.Join(intStrings, "-"))
		case map[string]string:
			// For custom headers, convert to JSON
			if jsonData, err := json.Marshal(v); err == nil {
				data.Set(key, string(jsonData))
			}
		default:
			// Try to convert to string
			data.Set(key, fmt.Sprintf("%v", v))
		}
	}
	
	// Create request
	req, err := http.NewRequest("POST", c.baseURL+endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "DomainVault-UptimeRobot/1.0")
	
	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	return body, nil
}

// checkAPIResponse checks if the API response indicates success
func checkAPIResponse(body []byte) error {
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}
	
	if apiResp.Stat != "ok" {
		if apiResp.Error != nil {
			return fmt.Errorf("API error (%s): %s", apiResp.Error.Type, apiResp.Error.Message)
		}
		return fmt.Errorf("API request failed with status: %s", apiResp.Stat)
	}
	
	return nil
}

// GetAccountDetails retrieves account information
func (c *Client) GetAccountDetails() (*Account, error) {
	body, err := c.makeRequest("/getAccountDetails", nil)
	if err != nil {
		return nil, err
	}
	
	if err := checkAPIResponse(body); err != nil {
		return nil, err
	}
	
	var response GetAccountDetailsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse account details response: %w", err)
	}
	
	return &response.Account, nil
}

// GetMonitors retrieves monitors based on the provided request
func (c *Client) GetMonitors(req *GetMonitorsRequest) ([]Monitor, error) {
	params := make(map[string]interface{})
	
	if req != nil {
		if len(req.Monitors) > 0 {
			params["monitors"] = req.Monitors
		}
		if len(req.Types) > 0 {
			params["types"] = req.Types
		}
		if len(req.Statuses) > 0 {
			params["statuses"] = req.Statuses
		}
		if len(req.CustomUptimeRatio) > 0 {
			params["custom_uptime_ratio"] = strings.Join(req.CustomUptimeRatio, "-")
		}
		if req.Logs > 0 {
			params["logs"] = req.Logs
			if req.LogsLimit > 0 {
				params["logs_limit"] = req.LogsLimit
			}
		}
		if req.ResponseTimes > 0 {
			params["response_times"] = req.ResponseTimes
			if req.ResponseTimesLimit > 0 {
				params["response_times_limit"] = req.ResponseTimesLimit
			}
		}
		if req.AlertContacts > 0 {
			params["alert_contacts"] = req.AlertContacts
		}
		if req.MaintenanceWindows > 0 {
			params["mwindows"] = req.MaintenanceWindows
		}
		if req.Timezone != "" {
			params["timezone"] = req.Timezone
		}
		if req.Offset > 0 {
			params["offset"] = req.Offset
		}
		if req.Limit > 0 {
			params["limit"] = req.Limit
		}
		if req.Search != "" {
			params["search"] = req.Search
		}
	}
	
	body, err := c.makeRequest("/getMonitors", params)
	if err != nil {
		return nil, err
	}
	
	if err := checkAPIResponse(body); err != nil {
		return nil, err
	}
	
	var response GetMonitorsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse monitors response: %w", err)
	}
	
	return response.Monitors, nil
}

// CreateMonitor creates a new monitor
func (c *Client) CreateMonitor(req *CreateMonitorRequest) (*CreateMonitorResponse, error) {
	params := map[string]interface{}{
		"friendly_name": req.FriendlyName,
		"url":          req.URL,
		"type":         int(req.Type),
	}
	
	if req.SubType != "" {
		params["sub_type"] = req.SubType
	}
	if req.Port != "" {
		params["port"] = req.Port
	}
	if req.KeywordType > 0 {
		params["keyword_type"] = int(req.KeywordType)
		params["keyword_value"] = req.KeywordValue
	}
	if req.HTTPUsername != "" {
		params["http_username"] = req.HTTPUsername
	}
	if req.HTTPPassword != "" {
		params["http_password"] = req.HTTPPassword
	}
	if req.Interval > 0 {
		params["interval"] = req.Interval
	} else {
		params["interval"] = DefaultInterval
	}
	if len(req.AlertContacts) > 0 {
		params["alert_contacts"] = strings.Join(req.AlertContacts, "-")
	}
	if len(req.CustomHTTPHeaders) > 0 {
		params["custom_http_headers"] = req.CustomHTTPHeaders
	}
	if req.CustomHTTPStatuses != "" {
		params["custom_http_statuses"] = req.CustomHTTPStatuses
	}
	if req.IgnoreSSLErrors > 0 {
		params["ignore_ssl_errors"] = req.IgnoreSSLErrors
	}
	
	body, err := c.makeRequest("/newMonitor", params)
	if err != nil {
		return nil, err
	}
	
	if err := checkAPIResponse(body); err != nil {
		return nil, err
	}
	
	var response CreateMonitorResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse create monitor response: %w", err)
	}
	
	return &response, nil
}

// UpdateMonitor updates an existing monitor
func (c *Client) UpdateMonitor(req *UpdateMonitorRequest) (*UpdateMonitorResponse, error) {
	params := map[string]interface{}{
		"id": req.ID,
	}
	
	if req.FriendlyName != "" {
		params["friendly_name"] = req.FriendlyName
	}
	if req.URL != "" {
		params["url"] = req.URL
	}
	if req.SubType != "" {
		params["sub_type"] = req.SubType
	}
	if req.Port != "" {
		params["port"] = req.Port
	}
	if req.KeywordType > 0 {
		params["keyword_type"] = int(req.KeywordType)
		params["keyword_value"] = req.KeywordValue
	}
	if req.HTTPUsername != "" {
		params["http_username"] = req.HTTPUsername
	}
	if req.HTTPPassword != "" {
		params["http_password"] = req.HTTPPassword
	}
	if req.Interval > 0 {
		params["interval"] = req.Interval
	}
	if req.Status > 0 {
		params["status"] = int(req.Status)
	}
	if len(req.AlertContacts) > 0 {
		params["alert_contacts"] = strings.Join(req.AlertContacts, "-")
	}
	if len(req.CustomHTTPHeaders) > 0 {
		params["custom_http_headers"] = req.CustomHTTPHeaders
	}
	if req.CustomHTTPStatuses != "" {
		params["custom_http_statuses"] = req.CustomHTTPStatuses
	}
	if req.IgnoreSSLErrors > 0 {
		params["ignore_ssl_errors"] = req.IgnoreSSLErrors
	}
	
	body, err := c.makeRequest("/editMonitor", params)
	if err != nil {
		return nil, err
	}
	
	if err := checkAPIResponse(body); err != nil {
		return nil, err
	}
	
	var response UpdateMonitorResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse update monitor response: %w", err)
	}
	
	return &response, nil
}

// DeleteMonitor deletes a monitor
func (c *Client) DeleteMonitor(monitorID int) (*DeleteMonitorResponse, error) {
	params := map[string]interface{}{
		"id": monitorID,
	}
	
	body, err := c.makeRequest("/deleteMonitor", params)
	if err != nil {
		return nil, err
	}
	
	if err := checkAPIResponse(body); err != nil {
		return nil, err
	}
	
	var response DeleteMonitorResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse delete monitor response: %w", err)
	}
	
	return &response, nil
}

// GetMonitorByName finds a monitor by its friendly name
func (c *Client) GetMonitorByName(name string) (*Monitor, error) {
	req := &GetMonitorsRequest{
		Search: name,
		Limit:  1,
	}
	
	monitors, err := c.GetMonitors(req)
	if err != nil {
		return nil, err
	}
	
	for _, monitor := range monitors {
		if monitor.FriendlyName == name {
			return &monitor, nil
		}
	}
	
	return nil, fmt.Errorf("monitor with name '%s' not found", name)
}

// GetMonitorsByURL finds monitors by URL
func (c *Client) GetMonitorsByURL(url string) ([]Monitor, error) {
	// Get all monitors and filter by URL since API doesn't support URL filtering
	monitors, err := c.GetMonitors(nil)
	if err != nil {
		return nil, err
	}
	
	var filtered []Monitor
	for _, monitor := range monitors {
		if monitor.URL == url || 
		   monitor.URL == "http://"+url || 
		   monitor.URL == "https://"+url {
			filtered = append(filtered, monitor)
		}
	}
	
	return filtered, nil
}

// TestConnection tests the API connection by retrieving account details
func (c *Client) TestConnection() error {
	_, err := c.GetAccountDetails()
	return err
}

// CreateHTTPMonitor creates an HTTP monitor for a domain
func (c *Client) CreateHTTPMonitor(domainName, friendlyName string, interval int, alertContacts []string) (*CreateMonitorResponse, error) {
	req := &CreateMonitorRequest{
		FriendlyName:  friendlyName,
		URL:          fmt.Sprintf("https://%s", domainName),
		Type:         MonitorTypeHTTP,
		Interval:     interval,
		AlertContacts: alertContacts,
	}
	
	// If HTTPS creation fails, try HTTP as fallback
	resp, err := c.CreateMonitor(req)
	if err != nil {
		// Try HTTP as fallback
		req.URL = fmt.Sprintf("http://%s", domainName)
		return c.CreateMonitor(req)
	}
	
	return resp, nil
}

// CreatePingMonitor creates a ping monitor for a domain
func (c *Client) CreatePingMonitor(domainName, friendlyName string, interval int, alertContacts []string) (*CreateMonitorResponse, error) {
	req := &CreateMonitorRequest{
		FriendlyName:  friendlyName,
		URL:          domainName, // Ping monitors use domain name without protocol
		Type:         MonitorTypePing,
		Interval:     interval,
		AlertContacts: alertContacts,
	}
	
	return c.CreateMonitor(req)
}

// CreateKeywordMonitor creates a keyword monitor for a domain
func (c *Client) CreateKeywordMonitor(domainName, friendlyName, keyword string, keywordExists bool, interval int, alertContacts []string) (*CreateMonitorResponse, error) {
	keywordType := KeywordTypeExists
	if !keywordExists {
		keywordType = KeywordTypeNotExists
	}
	
	req := &CreateMonitorRequest{
		FriendlyName:  friendlyName,
		URL:          fmt.Sprintf("https://%s", domainName),
		Type:         MonitorTypeKeyword,
		KeywordType:  keywordType,
		KeywordValue: keyword,
		Interval:     interval,
		AlertContacts: alertContacts,
	}
	
	// If HTTPS creation fails, try HTTP as fallback
	resp, err := c.CreateMonitor(req)
	if err != nil {
		req.URL = fmt.Sprintf("http://%s", domainName)
		return c.CreateMonitor(req)
	}
	
	return resp, nil
}

// PauseMonitor pauses a monitor
func (c *Client) PauseMonitor(monitorID int) error {
	req := &UpdateMonitorRequest{
		ID:     monitorID,
		Status: MonitorStatusPaused,
	}
	
	_, err := c.UpdateMonitor(req)
	return err
}

// ResumeMonitor resumes a paused monitor
func (c *Client) ResumeMonitor(monitorID int) error {
	req := &UpdateMonitorRequest{
		ID:     monitorID,
		Status: MonitorStatusNotCheckedYet, // This will start monitoring
	}

	_, err := c.UpdateMonitor(req)
	return err
}

// GetMonitorLogs retrieves logs for monitors
func (c *Client) GetMonitorLogs(req *GetMonitorLogsRequest) ([]MonitorLog, error) {
	params := make(map[string]interface{})
	
	if len(req.MonitorIDs) > 0 {
		params["monitors"] = req.MonitorIDs
	}
	if req.Limit > 0 {
		params["limit"] = req.Limit
	}
	if req.Offset > 0 {
		params["offset"] = req.Offset
	}
	if req.StartDate > 0 {
		params["start_date"] = req.StartDate
	}
	if req.EndDate > 0 {
		params["end_date"] = req.EndDate
	}
	if req.Timezone != "" {
		params["timezone"] = req.Timezone
	}
	
	body, err := c.makeRequest("/getLogs", params)
	if err != nil {
		return nil, err
	}
	
	if err := checkAPIResponse(body); err != nil {
		return nil, err
	}
	
	var response GetMonitorLogsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse monitor logs response: %w", err)
	}
	
	return response.Logs, nil
}

// GetMonitorStats gets detailed statistics for a monitor
func (c *Client) GetMonitorStats(monitorID int, customRanges ...string) (*Monitor, error) {
	req := &GetMonitorsRequest{
		Monitors:          []int{monitorID},
		ResponseTimes:     1,
		ResponseTimesLimit: 24, // Last 24 hours
		Logs:             1,
		LogsLimit:        10, // Last 10 events
		CustomUptimeRatio: customRanges,
	}
	
	monitors, err := c.GetMonitors(req)
	if err != nil {
		return nil, err
	}
	
	if len(monitors) == 0 {
		return nil, fmt.Errorf("monitor with ID %d not found", monitorID)
	}
	
	return &monitors[0], nil
}
