# With no retries

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.test-organization.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "oauth_token_id"
  service_provider = "github"
}

resource "tfe_workspace" "parent" {
  name           = "parent-ws"
  organization   = tfe_organization.test-organization.name
  queue_all_runs = false
  vcs_repo {
    branch         = "main"
    identifier     = "my-org-name/vcs-repository"
    oauth_token_id = tfe_oauth_client.test.oauth_token_id
  }
}

resource "tfe_workspace_run" "ws_run_parent" {
  workspace_id = tfe_workspace.parent.id

  apply {
    manual_confirm = false
    retry          = false
  }

  destroy {
    manual_confirm = false
    retry          = false
    wait_for_run   = true
  }
}
