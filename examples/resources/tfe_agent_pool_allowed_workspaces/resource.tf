# Basic usage
# In this example, the agent pool and workspace are connected through other resources that manage the agent pool permissions as well as the workspace execution mode. Notice that the `tfe_workspace_settings` uses the agent pool reference found in `tfe_agent_pool_allowed_workspaces` in order to create the permission to use the agent pool before assigning it.

// Ensure workspace and agent pool are create first
resource "tfe_workspace" "test-workspace" {
  name         = "my-workspace-name"
  organization = tfe_organization.example.name
}

resource "tfe_agent_pool" "test-agent-pool" {
  name                = "my-agent-pool-name"
  organization        = tfe_organization.example.name
  organization_scoped = false
}

// Ensure permissions are assigned second
resource "tfe_agent_pool_allowed_workspaces" "allowed" {
  agent_pool_id         = tfe_agent_pool.test-agent-pool.id
  allowed_workspace_ids = [tfe_workspace.test-workspace.id]
}

// Lastly, ensure the workspace agent execution is assigned last by
// referencing allowed_workspaces
resource "tfe_workspace_settings" "test-workspace-settings" {
  workspace_id   = tfe_workspace.test-workspace.id
  execution_mode = "agent"
  agent_pool_id  = tfe_agent_pool_allowed_workspaces.allowed.id
}
