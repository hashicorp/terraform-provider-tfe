# Basic usage

resource "tfe_workspace" "test-sourceable" {
  name         = "my-sourceable-workspace-name"
  organization = tfe_organization.example.name
}

resource "tfe_run_trigger" "test" {
  workspace_id  = tfe_workspace.example.id
  sourceable_id = tfe_workspace.test-sourceable.id
}
