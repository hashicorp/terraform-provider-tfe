# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Publish an immediate, site-wide announcement banner

variable "admin_token" {
  description = "An admin access token"
}

variable "hostname" {
  description = "The Terraform Enterprise hostname."
}

provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_admin_banner" "maintenance" {
  title    = "Scheduled maintenance"
  body     = "Terraform Enterprise will be <strong>unavailable</strong> on Sunday. See the <a href=\"https://status.example.com\">status page</a>."
  style    = "warning"
  audience = "all_users"
}
