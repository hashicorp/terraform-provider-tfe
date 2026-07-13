# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_ip_ranges" "addresses" {}

output "notifications_ips" {
  value = data.tfe_ip_ranges.addresses.notifications
}
