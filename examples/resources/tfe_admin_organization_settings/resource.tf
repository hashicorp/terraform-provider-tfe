# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

variable "token" {
  description = "An access token"
}

variable "admin_token" {
  description = "An admin access token"
}

variable "hostname" {
  description = "The HCP Terraform or Enterprise hostname."
  default     = "app.terraform.io"
}

provider "tfe" {
  hostname = var.hostname
  token    = var.token
}

provider "tfe" {
  alias    = "admin"
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_organization" "a-module-producer" {
  name  = "my-org"
  email = "admin@company.com"
}

resource "tfe_organization" "a-module-consumer" {
  name  = "my-other-org"
  email = "admin@company.com"
}

resource "tfe_admin_organization_settings" "test-settings" {
  provider              = tfe.admin
  organization          = tfe_organization.a-module-producer.name
  workspace_limit       = 15
  access_beta_tools     = false
  global_module_sharing = false
  module_sharing_consumer_organizations = [
    tfe_organization.a-module-consumer.name
  ]
}
