resource "uptimerobot_monitor" "domain_monitor" {
  for_each = {
    for key, domain in var.domains_to_monitor : key => domain
    if domain.enabled
  }

  friendly_name = each.value.name
  url           = each.value.url
  type          = each.value.monitor_type
  interval      = coalesce(each.value.interval, var.default_monitor_interval)
}
