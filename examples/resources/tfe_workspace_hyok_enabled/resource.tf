# Basic usage

resource "tfe_workspace" "example" {
  organization = "my-org-name"
  name         = "my-workspace"
}

resource "tfe_workspace_hyok_enabled" "example" {
  workspace_id = tfe_workspace.example.id
}
