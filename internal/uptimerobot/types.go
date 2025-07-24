package uptimerobot

import (
	"time"
)

// Monitor represents an UptimeRobot monitor
type Monitor struct {
	ID               int               `json:"id"`
	FriendlyName     string            `json:"friendly_name"`
	URL              string            `json:"url"`
	Type             MonitorType       `json:"type"`
	SubType          string            `json:"sub_type,omitempty"`
	KeywordType      KeywordType       `json:"keyword_type,omitempty"`
	KeywordValue     string            `json:"keyword_value,omitempty"`
	HTTPUsername     string            `json:"http_username,omitempty"`
	HTTPPassword     string            `json:"http_password,omitempty"`
	Port             string            `json:"port,omitempty"`
	Interval         int               `json:"interval"`
	Status           MonitorStatus     `json:"status"`
	CreateDatetime   int64             `json:"create_datetime"`
	MonitorGroup     int               `json:"monitor_group,omitempty"`
	IsGroupMain      int               `json:"is_group_main,omitempty"`
	ResponseTimes    []ResponseTime    `json:"response_times,omitempty"`
	Logs             []MonitorLog      `json:"logs,omitempty"`
	AlertContacts    []AlertContact    `json:"alert_contacts,omitempty"`
	MaintenanceWindows []MaintenanceWindow `json:"mwindows,omitempty"`
	CustomHTTPHeaders map[string]string `json:"custom_http_headers,omitempty"`
	CustomHTTPStatuses string           `json:"custom_http_statuses,omitempty"`
	SSLEnabled       int               `json:"ssl,omitempty"`
}

// MonitorType represents the type of monitor
type MonitorType int

const (
	MonitorTypeHTTP MonitorType = 1
	MonitorTypeKeyword MonitorType = 2
	MonitorTypePing MonitorType = 3
	MonitorTypePort MonitorType = 4
	MonitorTypeHeartbeat MonitorType = 5
)

// MonitorStatus represents the current status of a monitor
type MonitorStatus int

const (
	MonitorStatusPaused MonitorStatus = 0
	MonitorStatusNotCheckedYet MonitorStatus = 1
	MonitorStatusUp MonitorStatus = 2
	MonitorStatusSeemsDown MonitorStatus = 8
	MonitorStatusDown MonitorStatus = 9
)

// KeywordType for keyword monitoring
type KeywordType int

const (
	KeywordTypeExists KeywordType = 1
	KeywordTypeNotExists KeywordType = 2
)

// ResponseTime represents response time data
type ResponseTime struct {
	DatetimeISO interface{} `json:"datetime"` // Can be string or number
	Value       int         `json:"value"`
}

// MonitorLog represents monitor log entry
type MonitorLog struct {
	Type         int    `json:"type"`
	Datetime     int64  `json:"datetime"`
	DatetimeISO  string `json:"datetime_iso"`
	Duration     int    `json:"duration"`
	Reason       Reason `json:"reason"`
}

// Reason represents the reason for status change
type Reason struct {
	Code   interface{} `json:"code"` // Can be string or number depending on the API response
	Detail string      `json:"detail"`
}

// AlertContact represents an alert contact
type AlertContact struct {
	ID     string `json:"id"`
	Value  string `json:"value"`
	Type   int    `json:"type"`
	Status int    `json:"status"`
}

// MaintenanceWindow represents a maintenance window
type MaintenanceWindow struct {
	ID          int    `json:"id"`
	Type        int    `json:"type"`
	FriendlyName string `json:"friendly_name"`
	Value       string `json:"value"`
	StartTime   string `json:"start_time"`
	Duration    int    `json:"duration"`
	Status      int    `json:"status"`
}

// CreateMonitorRequest represents a request to create a new monitor
type CreateMonitorRequest struct {
	FriendlyName      string            `json:"friendly_name"`
	URL               string            `json:"url"`
	Type              MonitorType       `json:"type"`
	SubType           string            `json:"sub_type,omitempty"`
	Port              string            `json:"port,omitempty"`
	KeywordType       KeywordType       `json:"keyword_type,omitempty"`
	KeywordValue      string            `json:"keyword_value,omitempty"`
	HTTPUsername      string            `json:"http_username,omitempty"`
	HTTPPassword      string            `json:"http_password,omitempty"`
	Interval          int               `json:"interval,omitempty"` // 60, 120, 300, 600, 900, 1800, 3600
	AlertContacts     []string          `json:"alert_contacts,omitempty"`
	CustomHTTPHeaders map[string]string `json:"custom_http_headers,omitempty"`
	CustomHTTPStatuses string           `json:"custom_http_statuses,omitempty"`
	IgnoreSSLErrors   int               `json:"ignore_ssl_errors,omitempty"`
}

// UpdateMonitorRequest represents a request to update an existing monitor
type UpdateMonitorRequest struct {
	ID                int               `json:"id"`
	FriendlyName      string            `json:"friendly_name,omitempty"`
	URL               string            `json:"url,omitempty"`
	SubType           string            `json:"sub_type,omitempty"`
	Port              string            `json:"port,omitempty"`
	KeywordType       KeywordType       `json:"keyword_type,omitempty"`
	KeywordValue      string            `json:"keyword_value,omitempty"`
	HTTPUsername      string            `json:"http_username,omitempty"`
	HTTPPassword      string            `json:"http_password,omitempty"`
	Interval          int               `json:"interval,omitempty"`
	Status            MonitorStatus     `json:"status,omitempty"`
	AlertContacts     []string          `json:"alert_contacts,omitempty"`
	CustomHTTPHeaders map[string]string `json:"custom_http_headers,omitempty"`
	CustomHTTPStatuses string           `json:"custom_http_statuses,omitempty"`
	IgnoreSSLErrors   int               `json:"ignore_ssl_errors,omitempty"`
}

// DeleteMonitorRequest represents a request to delete a monitor
type DeleteMonitorRequest struct {
	ID int `json:"id"`
}

// GetMonitorsRequest represents a request to get monitors
type GetMonitorsRequest struct {
	Monitors          []int    `json:"monitors,omitempty"`          // Specific monitor IDs
	Types             []int    `json:"types,omitempty"`             // Monitor types to include
	Statuses          []int    `json:"statuses,omitempty"`          // Monitor statuses to include
	CustomUptimeRatio []string `json:"custom_uptime_ratio,omitempty"` // Custom time ranges
	Logs              int      `json:"logs,omitempty"`              // Include logs (0 or 1)
	LogsLimit         int      `json:"logs_limit,omitempty"`        // Limit log entries
	ResponseTimes     int      `json:"response_times,omitempty"`    // Include response times (0 or 1)
	ResponseTimesLimit int     `json:"response_times_limit,omitempty"` // Limit response time entries
	AlertContacts     int      `json:"alert_contacts,omitempty"`    // Include alert contacts (0 or 1)
	MaintenanceWindows int     `json:"mwindows,omitempty"`          // Include maintenance windows (0 or 1)
	Timezone          string   `json:"timezone,omitempty"`          // Timezone for logs/response times
	Offset            int      `json:"offset,omitempty"`            // Pagination offset
	Limit             int      `json:"limit,omitempty"`             // Pagination limit (max 50)
	Search            string   `json:"search,omitempty"`            // Search in monitor names
}

// APIResponse represents the base structure of UptimeRobot API responses
type APIResponse struct {
	Stat          string `json:"stat"`
	Error         *Error `json:"error,omitempty"`
	Pagination    *Pagination `json:"pagination,omitempty"`
	Account       *Account `json:"account,omitempty"`
}

// Error represents an API error
type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Pagination represents pagination information
type Pagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
}

// Account represents account information
type Account struct {
	Email            string `json:"email"`
	MonitorLimit     int    `json:"monitor_limit"`
	MonitorInterval  int    `json:"monitor_interval"`
	UpMonitors       int    `json:"up_monitors"`
	DownMonitors     int    `json:"down_monitors"`
	PausedMonitors   int    `json:"paused_monitors"`
}

// GetMonitorsResponse represents the response for getting monitors
type GetMonitorsResponse struct {
	APIResponse
	Monitors []Monitor `json:"monitors"`
}

// CreateMonitorResponse represents the response for creating a monitor
type CreateMonitorResponse struct {
	APIResponse
	Monitor struct {
		ID     int    `json:"id"`
		Status int    `json:"status"`
	} `json:"monitor"`
}

// UpdateMonitorResponse represents the response for updating a monitor
type UpdateMonitorResponse struct {
	APIResponse
	Monitor struct {
		ID int `json:"id"`
	} `json:"monitor"`
}

// DeleteMonitorResponse represents the response for deleting a monitor
type DeleteMonitorResponse struct {
	APIResponse
	Monitor struct {
		ID int `json:"id"`
	} `json:"monitor"`
}

// GetAccountDetailsResponse represents the response for getting account details
type GetAccountDetailsResponse struct {
	APIResponse
	Account Account `json:"account"`
}

// AlertContactType represents the type of alert contact
type AlertContactType int

const (
	AlertContactTypeSMS AlertContactType = 1
	AlertContactTypeEmail AlertContactType = 2
	AlertContactTypeTwitterDM AlertContactType = 3
	AlertContactTypeBoxcar AlertContactType = 4
	AlertContactTypeWebHook AlertContactType = 5
	AlertContactTypePushbullet AlertContactType = 6
	AlertContactTypePushover AlertContactType = 9
	AlertContactTypeHipChat AlertContactType = 10
	AlertContactTypeSlack AlertContactType = 11
)

// MonitorStats represents uptime statistics for a monitor
type MonitorStats struct {
	ID              int                    `json:"id"`
	FriendlyName    string                 `json:"friendly_name"`
	URL             string                 `json:"url"`
	Type            MonitorType            `json:"type"`
	Status          MonitorStatus          `json:"status"`
	UptimeRatio     float64               `json:"uptime_ratio"`
	ResponseTime    int                   `json:"response_time"`
	CreateDatetime  int64                 `json:"create_datetime"`
	CustomUptimeRatio map[string]float64  `json:"custom_uptime_ratio,omitempty"`
	Ranges          []UptimeRange         `json:"ranges,omitempty"`
}

// UptimeRange represents uptime data for a specific time range
type UptimeRange struct {
	StartDatetime int64   `json:"start_datetime"`
	EndDatetime   int64   `json:"end_datetime"`
	UptimeRatio   float64 `json:"uptime_ratio"`
	ResponseTime  int     `json:"response_time"`
}

// DomainMonitorConfig represents the configuration for domain monitoring
type DomainMonitorConfig struct {
	DomainID          string            // DomainVault domain ID
	DomainName        string            // Domain name to monitor
	MonitorType       MonitorType       // Type of monitoring (HTTP, HTTPS, Ping)
	Interval          int               // Check interval in seconds
	KeywordCheck      bool              // Enable keyword monitoring
	KeywordValue      string            // Keyword to search for
	KeywordType       KeywordType       // Keyword exists or not exists
	AlertContacts     []string          // Alert contact IDs
	CustomHeaders     map[string]string // Custom HTTP headers
	IgnoreSSLErrors   bool              // Ignore SSL certificate errors
	MaintenanceWindows []MaintenanceWindow // Maintenance windows
}

// MonitoringResult represents the result of a monitoring operation
type MonitoringResult struct {
	DomainID      string        `json:"domain_id"`
	DomainName    string        `json:"domain_name"`
	MonitorID     int           `json:"monitor_id,omitempty"`
	Success       bool          `json:"success"`
	Message       string        `json:"message"`
	HTTPStatus    int           `json:"http_status,omitempty"`
	ResponseTime  int           `json:"response_time,omitempty"`
	UptimeRatio   float64       `json:"uptime_ratio,omitempty"`
	LastChecked   time.Time     `json:"last_checked"`
	SSLStatus     string        `json:"ssl_status,omitempty"`
	Error         string        `json:"error,omitempty"`
}

// BulkMonitoringRequest represents a request to monitor multiple domains
type BulkMonitoringRequest struct {
	Domains          []string          `json:"domains"`
	MonitorType      MonitorType       `json:"monitor_type"`
	Interval         int               `json:"interval,omitempty"`
	AlertContacts    []string          `json:"alert_contacts,omitempty"`
	KeywordCheck     bool              `json:"keyword_check,omitempty"`
	KeywordValue     string            `json:"keyword_value,omitempty"`
	KeywordType      KeywordType       `json:"keyword_type,omitempty"`
	IgnoreSSLErrors  bool              `json:"ignore_ssl_errors,omitempty"`
}

// BulkMonitoringResponse represents the response for bulk monitoring setup
type BulkMonitoringResponse struct {
	Success      bool               `json:"success"`
	Message      string             `json:"message"`
	Results      []MonitoringResult `json:"results"`
	TotalCreated int                `json:"total_created"`
	TotalFailed  int                `json:"total_failed"`
}
