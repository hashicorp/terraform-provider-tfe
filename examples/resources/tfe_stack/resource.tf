# Create a stack with a VCS repository

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

data "tfe_agent_pool" "agent-pool" {
  name         = "my-example-agent-pool"
  organization = data.tfe_organization.organization.name
}

resource "tfe_stack" "test-stack" {
  name                = "my-stack"
  description         = "A Terraform Stack using two components with two environments"
  project_id          = data.tfe_organization.organization.default_project_id
  agent_pool_id       = data.tfe_agent_pool.agent-pool.id
  speculative_enabled = true

  vcs_repo {
    branch         = "main"
    identifier     = "my-github-org/stack-repo"
    oauth_token_id = tfe_oauth_client.test.oauth_token_id
  }
}
