# Basic usage

// Ensure project and agent pool are create first
resource "tfe_project" "test-project" {
  name         = "my-project-name"
  organization = tfe_organization.example.name
}

resource "tfe_agent_pool" "test-agent-pool" {
  name                = "my-agent-pool-name"
  organization        = tfe_organization.example.name
  organization_scoped = false
}

// Ensure permissions are assigned second
resource "tfe_agent_pool_allowed_projects" "allowed" {
  agent_pool_id       = tfe_agent_pool.test-agent-pool.id
  allowed_project_ids = [tfe_project.test-project.id]
}
