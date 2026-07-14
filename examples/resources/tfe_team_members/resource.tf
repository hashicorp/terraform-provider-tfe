# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_team_members" "test" {
  team_id   = tfe_team.test.id
  usernames = ["admin", "sander"]
}
