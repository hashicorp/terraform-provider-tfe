# Standard fixtures for example validation and live testing.
# These resources are injected alongside each example file by
# scripts/validate-examples.sh and TestAccExamples — they are NOT
# part of the rendered documentation.
#
# Do NOT reproduce these blocks in individual example files.
# Reference them directly using the label "example", e.g.:
#   organization = tfe_organization.example.name
#   workspace_id = tfe_workspace.example.id

resource "tfe_organization" "example" {
  name  = "my-org-name"
  email = "admin@example.com"
}

resource "tfe_workspace" "example" {
  name         = "my-workspace-name"
  organization = tfe_organization.example.name
}

resource "tfe_project" "example" {
  name         = "my-project-name"
  organization = tfe_organization.example.name
}

resource "tfe_agent_pool" "example" {
  name         = "my-agent-pool-name"
  organization = tfe_organization.example.name
}

