# With execution_mode of agent

resource "tfe_agent_pool_allowed_workspaces" "test" {
  agent_pool_id         = tfe_agent_pool.example.id
  allowed_workspace_ids = [tfe_workspace.example.id]
}

resource "tfe_workspace_settings" "test-settings" {
  workspace_id   = tfe_workspace.example.id
  agent_pool_id  = tfe_agent_pool_allowed_workspaces.test.agent_pool_id
  execution_mode = "agent"
}
