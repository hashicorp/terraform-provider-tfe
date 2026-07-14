# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "test-agent-pool" {
  name         = "my-agent-pool-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_oauth_client" "test-oauth-client" {
  organization     = tfe_organization.test-organization.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "my-vcs-provider-token"
  service_provider = "github"
}

resource "tfe_registry_module" "test-registry-module" {
  test_config {
    tests_enabled        = true
    agent_execution_mode = "agent"
    agent_pool_id        = tfe_agent_pool.test-agent-pool.id
  }

  vcs_repo {
    display_identifier = "my-org-name/terraform-provider-name"
    identifier         = "my-org-name/terraform-provider-name"
    oauth_token_id     = tfe_oauth_client.test-oauth-client.oauth_token_id
    branch             = "main"
  }
}
