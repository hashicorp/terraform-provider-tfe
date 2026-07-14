# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

locals {
  all_users = toset([
    "user1@hashicorp.com",
    "user2@hashicorp.com",
  ])
}

resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_organization_membership" "all_membership" {
  organization = "my-org-name"
  for_each     = local.all_users
  email        = each.key
}

resource "tfe_team_organization_members" "test" {
  team_id                     = tfe_team.test.id
  organization_membership_ids = [for member in local.all_users : tfe_organization_membership.all_membership[member].id]
}
