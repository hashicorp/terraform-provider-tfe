# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "my_agents" {
  name         = "my-agent-pool"
  organization = tfe_organization.test.name
}

# this section here for demonstration and is not necessary explicitly for use of tfe_project_settings
resource "tfe_organization_default_settings" "org_default" {
  organization = tfe_organization.test.name
  # this will end up being overwritten at the project level
  default_execution_mode = "agent"
  default_agent_pool_id  = tfe_agent_pool.my_agents.id
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
