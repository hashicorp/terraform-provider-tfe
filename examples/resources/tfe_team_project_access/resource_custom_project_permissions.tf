# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_team" "dev" {
  name         = "my-dev-team"
  organization = "my-org-name"
}

resource "tfe_project" "test" {
  name         = "myproject"
  organization = "my-org-name"
}

resource "tfe_team_project_access" "custom" {
  access     = "custom"
  team_id    = tfe_team.dev.id
  project_id = tfe_project.test.id

  project_access {
    settings      = "read"
    teams         = "none"
    variable_sets = "write"
  }
  workspace_access {
    state_versions   = "write"
    sentinel_mocks   = "none"
    runs             = "apply"
    variables        = "write"
    create           = true
    locking          = true
    move             = false
    delete           = false
    run_tasks        = false
    policy_overrides = true
  }
}
