# Basic usage

resource "tfe_workspace" "test" {
  organization = "my-org-name"
  name         = "my-workspace"
}

resource "tfe_workspace_hyok_enabled" "test" {
  workspace_id = tfe_workspace.test.id
}
