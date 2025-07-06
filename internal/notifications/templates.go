package notifications

import (
	"fmt"
	"strings"
	"time"

	"github.com/rusiqe/domainvault/internal/types"
)

// TemplateManager handles notification templates
type TemplateManager struct{}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{}
}

// RenderExpirationAlert renders expiration alert message
func (tm *TemplateManager) RenderExpirationAlert(domain types.Domain, daysUntilExpiry int) string {
	if daysUntilExpiry <= 0 {
		return fmt.Sprintf(`Domain %s has expired on %s. 
Please renew immediately to avoid losing the domain.
Provider: %s
Renewal Price: $%.2f
Auto-renew: %v`,
			domain.Name,
			domain.ExpiresAt.Format("January 2, 2006"),
			domain.Provider,
			getPrice(domain.RenewalPrice),
			domain.AutoRenew)
	}

	urgency := "soon"
	if daysUntilExpiry <= 7 {
		urgency = "very soon"
	} else if daysUntilExpiry <= 1 {
		urgency = "tomorrow"
	}

	return fmt.Sprintf(`Domain %s expires %s (%d days).
Expiration Date: %s
Provider: %s
Renewal Price: $%.2f
Auto-renew: %v

Please ensure renewal is processed in time.`,
		domain.Name,
		urgency,
		daysUntilExpiry,
		domain.ExpiresAt.Format("January 2, 2006"),
		domain.Provider,
		getPrice(domain.RenewalPrice),
		domain.AutoRenew)
}

// RenderStatusAlert renders status change alert message
func (tm *TemplateManager) RenderStatusAlert(domain types.Domain, previousStatus, currentStatus int) string {
	statusDesc := getStatusDescription(currentStatus)
	previousDesc := getStatusDescription(previousStatus)

	return fmt.Sprintf(`Domain %s status changed from %d (%s) to %d (%s).
Last checked: %s
Status message: %s

Please investigate if this is unexpected.`,
		domain.Name,
		previousStatus,
		previousDesc,
		currentStatus,
		statusDesc,
		formatTime(domain.LastStatusCheck),
		getStringPointer(domain.StatusMessage))
}

// RenderSyncFailureAlert renders sync failure alert message
func (tm *TemplateManager) RenderSyncFailureAlert(provider string, errorMsg string) string {
	return fmt.Sprintf(`Sync failed for provider %s.
Error: %s
Time: %s

This may affect domain data accuracy. Please check provider credentials and try again.`,
		provider,
		errorMsg,
		time.Now().Format("January 2, 2006 15:04:05"))
}

// RenderEmailAlert renders full HTML email for alerts
func (tm *TemplateManager) RenderEmailAlert(alert Alert) string {
	severityColor := getSeverityColorHex(alert.Severity)
	severityIcon := getSeverityIcon(alert.Severity)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>DomainVault Alert</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { background-color: %s; color: white; padding: 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; }
        .severity-badge { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: bold; text-transform: uppercase; margin-left: 10px; background-color: rgba(255,255,255,0.2); }
        .content { padding: 30px; }
        .alert-title { font-size: 20px; margin: 0 0 20px 0; color: #333; }
        .alert-message { font-size: 16px; line-height: 1.6; color: #666; margin-bottom: 30px; white-space: pre-line; }
        .details { background-color: #f8f9fa; padding: 20px; border-radius: 6px; border-left: 4px solid %s; }
        .details h3 { margin: 0 0 15px 0; color: #333; font-size: 16px; }
        .detail-item { margin: 8px 0; }
        .detail-label { font-weight: bold; color: #666; display: inline-block; width: 120px; }
        .detail-value { color: #333; }
        .footer { background-color: #f8f9fa; padding: 20px; text-align: center; color: #666; font-size: 14px; }
        .timestamp { color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>%s DomainVault Alert <span class="severity-badge">%s</span></h1>
        </div>
        <div class="content">
            <h2 class="alert-title">%s</h2>
            <div class="alert-message">%s</div>
            <div class="details">
                <h3>Alert Details</h3>
                <div class="detail-item">
                    <span class="detail-label">Alert Type:</span>
                    <span class="detail-value">%s</span>
                </div>
                <div class="detail-item">
                    <span class="detail-label">Severity:</span>
                    <span class="detail-value">%s</span>
                </div>
                <div class="detail-item">
                    <span class="detail-label">Triggered By:</span>
                    <span class="detail-value">%s</span>
                </div>
                <div class="detail-item">
                    <span class="detail-label">Created At:</span>
                    <span class="detail-value">%s</span>
                </div>
                %s
            </div>
        </div>
        <div class="footer">
            <p>This alert was generated by DomainVault monitoring system.</p>
            <p class="timestamp">Alert ID: %s</p>
        </div>
    </div>
</body>
</html>`,
		severityColor,
		severityColor,
		severityIcon,
		strings.ToUpper(string(alert.Severity)),
		alert.Title,
		alert.Message,
		string(alert.Type),
		string(alert.Severity),
		alert.TriggeredBy,
		alert.CreatedAt.Format("January 2, 2006 15:04:05 MST"),
		tm.renderAlertData(alert.Data),
		alert.ID,
	)

	return html
}

// renderAlertData renders additional alert data as HTML
func (tm *TemplateManager) renderAlertData(data map[string]interface{}) string {
	if len(data) == 0 {
		return ""
	}

	var items []string
	for key, value := range data {
		// Skip internal fields and format user-friendly labels
		if strings.HasPrefix(key, "_") {
			continue
		}

		label := tm.formatLabel(key)
		formattedValue := tm.formatValue(value)
		
		items = append(items, fmt.Sprintf(`
                <div class="detail-item">
                    <span class="detail-label">%s:</span>
                    <span class="detail-value">%s</span>
                </div>`, label, formattedValue))
	}

	if len(items) > 0 {
		return `<h3 style="margin-top: 20px;">Additional Information</h3>` + strings.Join(items, "")
	}

	return ""
}

// formatLabel converts snake_case to Title Case
func (tm *TemplateManager) formatLabel(key string) string {
	words := strings.Split(key, "_")
	for i, word := range words {
		words[i] = strings.Title(word)
	}
	return strings.Join(words, " ")
}

// formatValue formats interface{} values for display
func (tm *TemplateManager) formatValue(value interface{}) string {
	switch v := value.(type) {
	case time.Time:
		return v.Format("January 2, 2006 15:04:05")
	case *time.Time:
		if v != nil {
			return v.Format("January 2, 2006 15:04:05")
		}
		return "Never"
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%.2f", v)
	case *float64:
		if v != nil {
			if *v == float64(int64(*v)) {
				return fmt.Sprintf("%.0f", *v)
			}
			return fmt.Sprintf("%.2f", *v)
		}
		return "N/A"
	case bool:
		if v {
			return "Yes"
		}
		return "No"
	case string:
		if v == "" {
			return "N/A"
		}
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Helper functions
func getPrice(price *float64) float64 {
	if price != nil {
		return *price
	}
	return 0.0
}

func getStringPointer(s *string) string {
	if s != nil {
		return *s
	}
	return "N/A"
}

func formatTime(t *time.Time) string {
	if t != nil {
		return t.Format("January 2, 2006 15:04:05")
	}
	return "Never"
}

func getStatusDescription(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "Success"
	case status >= 300 && status < 400:
		return "Redirect"
	case status >= 400 && status < 500:
		return "Client Error"
	case status >= 500:
		return "Server Error"
	default:
		return "Unknown"
	}
}

func getSeverityColorHex(severity AlertSeverity) string {
	switch severity {
	case SeverityCritical:
		return "#dc3545"
	case SeverityHigh:
		return "#fd7e14"
	case SeverityMedium:
		return "#ffc107"
	case SeverityLow:
		return "#28a745"
	default:
		return "#6c757d"
	}
}

func getSeverityIcon(severity AlertSeverity) string {
	switch severity {
	case SeverityCritical:
		return "🚨"
	case SeverityHigh:
		return "⚠️"
	case SeverityMedium:
		return "⚡"
	case SeverityLow:
		return "ℹ️"
	default:
		return "📢"
	}
}
