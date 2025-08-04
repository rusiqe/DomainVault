# Variables for UptimeRobot Terraform configuration

variable "uptimerobot_api_key" {
  description = "UptimeRobot API key"
  type        = string
  sensitive   = true
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "prod"
}

variable "default_monitor_interval" {
  description = "Default monitoring interval in seconds"
  type        = number
  default     = 300
  
  validation {
    condition = contains([60, 120, 300, 600, 900, 1800, 3600], var.default_monitor_interval)
    error_message = "Monitor interval must be one of: 60, 120, 300, 600, 900, 1800, 3600 seconds."
  }
}

variable "alert_contact_emails" {
  description = "List of alert contact email friendly names"
  type        = list(string)
  default     = []
}

variable "domains_to_monitor" {
  description = "Map of domains to monitor with their configurations"
  type = map(object({
    name          = string
    url           = string
    monitor_type  = string
    interval      = optional(number)
    keyword       = optional(string)
    keyword_type  = optional(string)
    enabled       = optional(bool, true)
    ssl_check     = optional(bool, true)
    port          = optional(number)
  }))
  default = {}
}

variable "maintenance_windows" {
  description = "Maintenance windows configuration"
  type = map(object({
    friendly_name = string
    type         = number
    value        = string
    start_time   = string
    duration     = number
  }))
  default = {}
}

variable "notification_groups" {
  description = "Notification groups for different domain categories"
  type = map(object({
    name            = string
    alert_contacts  = list(string)
    description     = optional(string)
  }))
  default = {
    critical = {
      name = "Critical Domains"
      alert_contacts = []
      description = "High priority domains requiring immediate attention"
    }
    standard = {
      name = "Standard Domains" 
      alert_contacts = []
      description = "Regular monitoring domains"
    }
  }
}

variable "auto_ssl_monitoring" {
  description = "Enable automatic SSL certificate monitoring"
  type        = bool
  default     = true
}

variable "response_time_threshold" {
  description = "Response time threshold in milliseconds for alerts"
  type        = number
  default     = 5000
}

variable "uptime_threshold" {
  description = "Uptime percentage threshold for alerts"
  type        = number
  default     = 99.0
  
  validation {
    condition = var.uptime_threshold >= 0 && var.uptime_threshold <= 100
    error_message = "Uptime threshold must be between 0 and 100."
  }
}
