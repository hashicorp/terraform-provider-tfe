# Basic usage

resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "my_agents" {
  name         = "agent_smiths"
  organization = tfe_organization.test.name
}

resource "tfe_organization_default_settings" "org_default" {
  organization           = tfe_organization.test.name
  default_execution_mode = "agent"
  default_agent_pool_id  = tfe_agent_pool.my_agents.id
}

resource "tfe_workspace" "my_workspace" {
  name = "my-workspace"
  # This workspace will use the org defaults, and will report those defaults as
  # the values of its corresponding attributes. Use depends_on to get accurate
  # values immediately, and to ensure reliable behavior of tfe_workspace_run.
  depends_on = [tfe_organization_default_settings.org_default]
}
