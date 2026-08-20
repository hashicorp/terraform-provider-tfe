# Basic usage

resource "tfe_workspace_settings" "test-settings" {
  workspace_id   = tfe_workspace.example.id
  execution_mode = "local"
}
