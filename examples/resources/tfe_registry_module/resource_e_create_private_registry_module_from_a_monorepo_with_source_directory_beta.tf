# Create private registry module from a monorepo with source_directory (BETA)

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test-oauth-client" {
  organization     = tfe_organization.test-organization.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "my-vcs-provider-token"
  service_provider = "github"
}

resource "tfe_registry_module" "monorepo-module" {
  organization    = tfe_organization.test-organization.name
  name            = "vpc"
  module_provider = "aws"

  vcs_repo {
    display_identifier = "my-org-name/private-modules"
    identifier         = "my-org-name/private-modules"
    oauth_token_id     = tfe_oauth_client.test-oauth-client.oauth_token_id
    source_directory   = "modules/vpc"
  }
}
