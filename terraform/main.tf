# Terraform configuration for UptimeRobot monitoring
terraform {
  required_version = ">= 1.0"
  required_providers {
    uptimerobot = {
      source  = "vexxhost/uptimerobot"
    }
  }
}

# Configure the UptimeRobot Provider
provider "uptimerobot" {
  api_key = var.uptimerobot_api_key
}

# Data source to get existing alert contacts (commented out until contacts are created)
# data "uptimerobot_alert_contact" "email" {
#   for_each = toset(var.alert_contact_emails)
#   
#   friendly_name = each.value
# }

# Data source to get account details
data "uptimerobot_account" "main" {}

# Local values for common configurations
locals {
  # Default monitor settings
  default_interval = var.default_monitor_interval
  
  # Alert contact IDs from data sources (disabled for now)
  # alert_contact_ids = [
  #   for contact in data.uptimerobot_alert_contact.email : contact.id
  # ]
  
  # Domain configurations from external data or variables
  domains_to_monitor = var.domains_to_monitor
  
  # Common tags for all monitors
  common_tags = {
    managed_by = "terraform"
    project    = "domainvault"
    environment = var.environment
  }
}
