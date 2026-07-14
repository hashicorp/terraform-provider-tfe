# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_team" "admin" {
  name         = "my-admin-team"
  organization = "my-org-name"
}

resource "tfe_project" "test" {
  name         = "myproject"
  organization = "my-org-name"
}

resource "tfe_team_project_access" "admin" {
  access     = "admin"
  team_id    = tfe_team.admin.id
  project_id = tfe_project.test.id
}
