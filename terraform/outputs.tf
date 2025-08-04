# Outputs for UptimeRobot Terraform configuration

output "account_info" {
  description = "UptimeRobot account information"
  value = {
    email           = data.uptimerobot_account.main.email
    monitor_limit   = data.uptimerobot_account.main.monitor_limit
    monitor_interval = data.uptimerobot_account.main.monitor_interval
    up_monitors     = data.uptimerobot_account.main.up_monitors
    down_monitors   = data.uptimerobot_account.main.down_monitors
    paused_monitors = data.uptimerobot_account.main.paused_monitors
  }
}

# output "alert_contacts" {
#   description = "Alert contact IDs"
#   value = {
#     for name, contact in data.uptimerobot_alert_contact.email : 
#     name => {
#       id    = contact.id
#       type  = contact.type
#       value = contact.value
#     }
#   }
# }

output "monitors" {
  description = "Created monitor information"
  value = {
    for key, monitor in uptimerobot_monitor.domain_monitor :
    key => {
      id            = monitor.id
      friendly_name = monitor.friendly_name
      url           = monitor.url
      type          = monitor.type
      status        = monitor.status
      interval      = monitor.interval
    }
  }
}

output "monitor_urls" {
  description = "List of all monitored URLs"
  value = [
    for monitor in uptimerobot_monitor.domain_monitor : monitor.url
  ]
}

output "monitor_count" {
  description = "Total number of monitors created"
  value = length(uptimerobot_monitor.domain_monitor)
}

output "monitoring_summary" {
  description = "Summary of monitoring setup"
  value = {
    total_monitors    = length(uptimerobot_monitor.domain_monitor)
    account_limit     = data.uptimerobot_account.main.monitor_limit
    remaining_slots   = data.uptimerobot_account.main.monitor_limit - length(uptimerobot_monitor.domain_monitor)
    environment       = var.environment
  }
}
