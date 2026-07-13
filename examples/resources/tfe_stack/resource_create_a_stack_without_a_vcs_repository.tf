# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

variable "github_token" {
  description = "An access token for github"
}

resource "tfe_oauth_client" "test" {
  organization     = "my-example-org"
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = var.github_token
  service_provider = "github"
}

data "tfe_organization" "organization" {
  name = "my-example-org"
}

resource "tfe_stack" "test-stack" {
  name        = "my-stack"
  description = "A Terraform Stack using two components with two environments"
  project_id  = data.tfe_organization.organization.default_project_id
}
