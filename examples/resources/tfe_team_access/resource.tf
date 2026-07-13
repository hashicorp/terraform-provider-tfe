# Basic usage

resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = "my-org-name"
}

resource "tfe_team_access" "test" {
  team_id      = tfe_team.test.id
  workspace_id = tfe_workspace.test.id

  permissions {
    runs              = "plan"
    variables         = "read"
    state_versions    = "read-outputs"
    sentinel_mocks    = "none"
    workspace_locking = false
    run_tasks         = false
    policy_overrides  = true
  }
}
