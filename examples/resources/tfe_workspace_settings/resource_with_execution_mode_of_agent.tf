# With execution_mode of agent

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "test-agent-pool" {
  name         = "my-agent-pool-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_agent_pool_allowed_workspaces" "test" {
  agent_pool_id         = tfe_agent_pool.test-agent-pool.id
  allowed_workspace_ids = [tfe_workspace.test.id]
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_workspace_settings" "test-settings" {
  workspace_id   = tfe_workspace.test.id
  agent_pool_id  = tfe_agent_pool_allowed_workspaces.test.agent_pool_id
  execution_mode = "agent"
}
