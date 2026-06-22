data "tfe_ip_ranges" "addresses" {}

output "notifications_ips" {
  value = data.tfe_ip_ranges.addresses.notifications
}