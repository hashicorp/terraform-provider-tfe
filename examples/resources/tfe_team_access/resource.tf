# Basic usage

resource "tfe_team" "test" {
  name         = "access-team"
  organization = "my-org-name"
}

resource "tfe_workspace" "test" {
  name         = "access-workspace"
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
