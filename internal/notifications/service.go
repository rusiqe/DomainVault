package notifications

import (
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/rusiqe/domainvault/internal/types"
)

// NotificationService handles all notification types
type NotificationService struct {
	emailConfig   EmailConfig
	slackConfig   SlackConfig
	webhookConfig WebhookConfig
	templates     *TemplateManager
}

// EmailConfig contains SMTP configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	Username     string
	Password     string
	FromAddress  string
	FromName     string
	Enabled      bool
}

// SlackConfig contains Slack webhook configuration
type SlackConfig struct {
	WebhookURL string
	Channel    string
	Username   string
	Enabled    bool
}

// WebhookConfig contains custom webhook configuration
type WebhookConfig struct {
	URLs    []string
	Secret  string
	Enabled bool
}

// AlertType represents different types of alerts
type AlertType string

const (
	AlertExpiringSoon   AlertType = "expiring_soon"
	AlertExpired        AlertType = "expired"
	AlertStatusDown     AlertType = "status_down"
	AlertDNSChanged     AlertType = "dns_changed"
	AlertSyncFailed     AlertType = "sync_failed"
	AlertBulkOperation  AlertType = "bulk_operation"
	AlertSecurity       AlertType = "security"
)

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	SeverityLow      AlertSeverity = "low"
	SeverityMedium   AlertSeverity = "medium"
	SeverityHigh     AlertSeverity = "high"
	SeverityCritical AlertSeverity = "critical"
)

// Alert represents a notification alert
type Alert struct {
	ID          string                 `json:"id"`
	Type        AlertType              `json:"type"`
	Severity    AlertSeverity          `json:"severity"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
	TriggeredBy string                 `json:"triggered_by"`
}

// NotificationChannel represents a notification delivery method
type NotificationChannel string

const (
	ChannelEmail   NotificationChannel = "email"
	ChannelSlack   NotificationChannel = "slack"
	ChannelWebhook NotificationChannel = "webhook"
	ChannelSMS     NotificationChannel = "sms"
)

// NotificationRule defines when and how to send notifications
type NotificationRule struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	AlertTypes  []AlertType           `json:"alert_types"`
	Severities  []AlertSeverity       `json:"severities"`
	Channels    []NotificationChannel `json:"channels"`
	Recipients  []string              `json:"recipients"`
	Conditions  map[string]interface{} `json:"conditions"`
	Enabled     bool                  `json:"enabled"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// NewNotificationService creates a new notification service
func NewNotificationService(emailConfig EmailConfig, slackConfig SlackConfig, webhookConfig WebhookConfig) *NotificationService {
	return &NotificationService{
		emailConfig:   emailConfig,
		slackConfig:   slackConfig,
		webhookConfig: webhookConfig,
		templates:     NewTemplateManager(),
	}
}

// SendAlert sends an alert through configured channels
func (ns *NotificationService) SendAlert(alert Alert, rules []NotificationRule) error {
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// Check if rule matches this alert
		if !ns.matchesRule(alert, rule) {
			continue
		}

		// Send through each configured channel
		for _, channel := range rule.Channels {
			switch channel {
			case ChannelEmail:
				if ns.emailConfig.Enabled {
					if err := ns.sendEmailAlert(alert, rule.Recipients); err != nil {
						log.Printf("Failed to send email alert: %v", err)
					}
				}
			case ChannelSlack:
				if ns.slackConfig.Enabled {
					if err := ns.sendSlackAlert(alert); err != nil {
						log.Printf("Failed to send Slack alert: %v", err)
					}
				}
			case ChannelWebhook:
				if ns.webhookConfig.Enabled {
					if err := ns.sendWebhookAlert(alert); err != nil {
						log.Printf("Failed to send webhook alert: %v", err)
					}
				}
			}
		}
	}

	return nil
}

// CreateExpirationAlert creates alerts for expiring domains
func (ns *NotificationService) CreateExpirationAlert(domain types.Domain, daysUntilExpiry int) Alert {
	severity := SeverityMedium
	alertType := AlertExpiringSoon

	if daysUntilExpiry <= 0 {
		severity = SeverityCritical
		alertType = AlertExpired
	} else if daysUntilExpiry <= 7 {
		severity = SeverityHigh
	} else if daysUntilExpiry <= 30 {
		severity = SeverityMedium
	} else {
		severity = SeverityLow
	}

	title := fmt.Sprintf("Domain %s expires in %d days", domain.Name, daysUntilExpiry)
	if daysUntilExpiry <= 0 {
		title = fmt.Sprintf("Domain %s has expired", domain.Name)
	}

	message := ns.templates.RenderExpirationAlert(domain, daysUntilExpiry)

	return Alert{
		ID:       fmt.Sprintf("exp_%s_%d", domain.ID, time.Now().Unix()),
		Type:     alertType,
		Severity: severity,
		Title:    title,
		Message:  message,
		Data: map[string]interface{}{
			"domain_id":          domain.ID,
			"domain_name":        domain.Name,
			"provider":           domain.Provider,
			"expires_at":         domain.ExpiresAt,
			"days_until_expiry":  daysUntilExpiry,
			"renewal_price":      domain.RenewalPrice,
			"auto_renew":         domain.AutoRenew,
		},
		CreatedAt:   time.Now(),
		TriggeredBy: "expiration_monitor",
	}
}

// CreateStatusAlert creates alerts for domain status changes
func (ns *NotificationService) CreateStatusAlert(domain types.Domain, previousStatus, currentStatus int) Alert {
	severity := SeverityMedium
	if currentStatus >= 500 {
		severity = SeverityHigh
	} else if currentStatus >= 400 {
		severity = SeverityMedium
	}

	title := fmt.Sprintf("Domain %s status changed to %d", domain.Name, currentStatus)
	message := ns.templates.RenderStatusAlert(domain, previousStatus, currentStatus)

	return Alert{
		ID:       fmt.Sprintf("status_%s_%d", domain.ID, time.Now().Unix()),
		Type:     AlertStatusDown,
		Severity: severity,
		Title:    title,
		Message:  message,
		Data: map[string]interface{}{
			"domain_id":        domain.ID,
			"domain_name":      domain.Name,
			"previous_status":  previousStatus,
			"current_status":   currentStatus,
			"status_message":   domain.StatusMessage,
			"last_check":       domain.LastStatusCheck,
		},
		CreatedAt:   time.Now(),
		TriggeredBy: "status_monitor",
	}
}

// CreateSyncFailureAlert creates alerts for sync failures
func (ns *NotificationService) CreateSyncFailureAlert(provider string, errorMsg string) Alert {
	return Alert{
		ID:       fmt.Sprintf("sync_%s_%d", provider, time.Now().Unix()),
		Type:     AlertSyncFailed,
		Severity: SeverityHigh,
		Title:    fmt.Sprintf("Sync failed for provider %s", provider),
		Message:  ns.templates.RenderSyncFailureAlert(provider, errorMsg),
		Data: map[string]interface{}{
			"provider":    provider,
			"error":       errorMsg,
			"sync_time":   time.Now(),
		},
		CreatedAt:   time.Now(),
		TriggeredBy: "sync_service",
	}
}

// matchesRule checks if an alert matches a notification rule
func (ns *NotificationService) matchesRule(alert Alert, rule NotificationRule) bool {
	// Check alert type
	typeMatches := false
	for _, alertType := range rule.AlertTypes {
		if alertType == alert.Type {
			typeMatches = true
			break
		}
	}
	if !typeMatches && len(rule.AlertTypes) > 0 {
		return false
	}

	// Check severity
	severityMatches := false
	for _, severity := range rule.Severities {
		if severity == alert.Severity {
			severityMatches = true
			break
		}
	}
	if !severityMatches && len(rule.Severities) > 0 {
		return false
	}

	// Additional condition checks could be added here
	return true
}

// sendEmailAlert sends an alert via email
func (ns *NotificationService) sendEmailAlert(alert Alert, recipients []string) error {
	if !ns.emailConfig.Enabled || len(recipients) == 0 {
		return fmt.Errorf("email not configured or no recipients")
	}

	subject := fmt.Sprintf("[DomainVault %s] %s", strings.ToUpper(string(alert.Severity)), alert.Title)
	body := ns.templates.RenderEmailAlert(alert)

	// Setup authentication
	auth := smtp.PlainAuth("", ns.emailConfig.Username, ns.emailConfig.Password, ns.emailConfig.SMTPHost)

	// Compose message
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", 
		strings.Join(recipients, ","), subject, body))

	// Send email
	addr := fmt.Sprintf("%s:%d", ns.emailConfig.SMTPHost, ns.emailConfig.SMTPPort)
	return smtp.SendMail(addr, auth, ns.emailConfig.FromAddress, recipients, msg)
}

// sendSlackAlert sends an alert via Slack webhook
func (ns *NotificationService) sendSlackAlert(alert Alert) error {
	if !ns.slackConfig.Enabled {
		return fmt.Errorf("Slack not configured")
	}

	// Create Slack message payload
	payload := map[string]interface{}{
		"channel":  ns.slackConfig.Channel,
		"username": ns.slackConfig.Username,
		"text":     alert.Title,
		"attachments": []map[string]interface{}{
			{
				"color":  ns.getSeverityColor(alert.Severity),
				"fields": []map[string]interface{}{
					{
						"title": "Alert Type",
						"value": string(alert.Type),
						"short": true,
					},
					{
						"title": "Severity",
						"value": string(alert.Severity),
						"short": true,
					},
					{
						"title": "Message",
						"value": alert.Message,
						"short": false,
					},
				},
				"timestamp": alert.CreatedAt.Unix(),
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	// Send to Slack webhook (implementation would use HTTP client)
	log.Printf("Would send to Slack: %s", string(jsonPayload))
	return nil
}

// sendWebhookAlert sends an alert via custom webhook
func (ns *NotificationService) sendWebhookAlert(alert Alert) error {
	if !ns.webhookConfig.Enabled {
		return fmt.Errorf("webhooks not configured")
	}

	// Create webhook payload
	payload := map[string]interface{}{
		"alert":      alert,
		"timestamp":  time.Now().Unix(),
		"signature":  ns.generateWebhookSignature(alert),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Send to each configured webhook URL (implementation would use HTTP client)
	for _, url := range ns.webhookConfig.URLs {
		log.Printf("Would send to webhook %s: %s", url, string(jsonPayload))
	}

	return nil
}

// getSeverityColor returns Slack color for alert severity
func (ns *NotificationService) getSeverityColor(severity AlertSeverity) string {
	switch severity {
	case SeverityCritical:
		return "danger"
	case SeverityHigh:
		return "warning"
	case SeverityMedium:
		return "#ff9500"
	case SeverityLow:
		return "good"
	default:
		return "#cccccc"
	}
}

// generateWebhookSignature generates HMAC signature for webhook security
func (ns *NotificationService) generateWebhookSignature(alert Alert) string {
	// Implementation would use HMAC-SHA256 with webhook secret
	return "webhook_signature_placeholder"
}
