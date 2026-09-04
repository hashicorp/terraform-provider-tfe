# Basic usage

resource "tfe_agent_pool" "test-agent-pool" {
  name                = "my-agent-pool-name"
  organization        = tfe_organization.example.name
  organization_scoped = true
}
