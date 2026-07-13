# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_organization_membership" "test" {
  organization = "my-org-name"
  email        = "example@hashicorp.com"
}

resource "tfe_organization_membership" "sample" {
  organization = "my-org-name"
  email        = "sample@hashicorp.com"
}

resource "tfe_team_organization_members" "test" {
  team_id = tfe_team.test.id
  organization_membership_ids = [
    tfe_organization_membership.test.id,
    tfe_organization_membership.sample.id
  ]
}
