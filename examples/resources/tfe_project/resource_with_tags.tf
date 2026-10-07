# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# With tags

resource "tfe_project" "test" {
  organization = tfe_organization.example.name
  name         = "projectname"
  tags = {
    cost_center = "infrastructure"
    team        = "platform"
  }
}
