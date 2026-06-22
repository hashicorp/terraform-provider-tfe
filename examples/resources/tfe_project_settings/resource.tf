resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"

  # this will end up being overwritten at the project level
  default_execution_mode = "remote"
}

resource "tfe_agent_pool" "my_agents" {
  name         = "my-agent-pool"
  organization = tfe_organization.test.name
}

resource "tfe_project" "my_project" {
  name         = "my-project"
  organization = tfe_organization.test.name
}

resource "tfe_project_settings" "my_project_settings" {
  project_id = tfe_project.my_project.id

  # workspaces in this project will use agent execution mode by default,
  # and will use the specified agent pool.
  default_execution_mode = "agent"
  default_agent_pool_id  = tfe_agent_pool.my_agents.id
}
