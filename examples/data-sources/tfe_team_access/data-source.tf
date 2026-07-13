# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_team_access" "test" {
  team_id      = "my-team-id"
  workspace_id = "my-workspace-id"
}
