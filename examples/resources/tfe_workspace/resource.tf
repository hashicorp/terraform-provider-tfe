# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_workspace" "test" {
  name         = "user-workspace"
  organization = tfe_organization.example.name
  tags = {
    environment = "prod"
    team_owner  = "my-team"
  }
}
