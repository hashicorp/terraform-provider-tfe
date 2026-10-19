# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Manage admin customization settings for a Terraform Enterprise instance.

variable "admin_token" {
  description = "An admin access token"
}

variable "hostname" {
  description = "The HCP Terraform or Enterprise hostname."
  default     = "app.terraform.io"
}

provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_admin_customization_settings" "this" {
  support_email              = "support@example.com"
  internal_support_url       = "https://internal.example.com/support"
  internal_documentation_url = "https://internal.example.com/docs"
  internal_tutorials_url     = "https://internal.example.com/tutorials"
  login_help                 = "<p>Contact your administrator for help signing in.</p>"
  footer                     = "<p>© Example Corp</p>"
  error                      = "<p>Something went wrong. Please contact support.</p>"
  new_user                   = "<p>Welcome to Terraform Enterprise!</p>"

  avatar_sources = {
    gravatar = true
  }
}
