# Basic usage (VCS-based policy set)

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.test-organization.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "oauth_token_id"
  service_provider = "github"
}

resource "tfe_policy_set" "test" {
  name                = "my-policy-set"
  description         = "A brand new policy set"
  organization        = "my-org-name"
  kind                = "sentinel"
  agent_enabled       = "true"
  policy_tool_version = "0.24.1"
  # Top-level policy set argument that applies when vcs_repo is configured.
  policy_update_patterns = ["**/*.sentinel", "policies/**/*.hcl"]
  policies_path          = "policies/my-policy-set"
  workspace_ids          = [tfe_workspace.test.id]

  vcs_repo {
    identifier         = "my-org-name/my-policy-set-repository"
    branch             = "main"
    ingress_submodules = false
    oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  }
}
