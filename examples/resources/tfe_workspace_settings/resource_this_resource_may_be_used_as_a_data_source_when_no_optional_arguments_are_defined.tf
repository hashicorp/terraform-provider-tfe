data "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = "my-org-name"
}

resource "tfe_workspace_settings" "test" {
  workspace_id = data.tfe_workspace.test.id
}

output "workspace-explicit-local-execution" {
  value = alltrue([
    tfe_workspace_settings.test.execution_mode == "local",
    tfe_workspace_settings.test.overwrites[0]["execution_mode"]
  ])
}
