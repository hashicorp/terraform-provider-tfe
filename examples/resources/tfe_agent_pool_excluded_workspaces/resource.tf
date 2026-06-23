resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

// Ensure workspace and agent pool are create first
resource "tfe_workspace" "test-workspace" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_agent_pool" "test-agent-pool" {
  name                = "my-agent-pool-name"
  organization        = tfe_organization.test-organization.name
  organization_scoped = false
}

// Ensure permissions are assigned second
resource "tfe_agent_pool_excluded_workspaces" "excluded" {
  agent_pool_id          = tfe_agent_pool.test-agent-pool.id
  excluded_workspace_ids = [tfe_workspace.test-workspace.id]
}
