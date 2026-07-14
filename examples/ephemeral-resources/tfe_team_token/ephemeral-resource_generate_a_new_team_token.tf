# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_team" "example" {
  organization = "my-org-name"
  name         = "my-team-name"
}

ephemeral "tfe_team_token" "example" {
  team_id = tfe_team.example.id
}
