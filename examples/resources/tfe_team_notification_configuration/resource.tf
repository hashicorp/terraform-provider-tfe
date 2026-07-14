# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = tfe_organization.test.id
}

resource "tfe_team_notification_configuration" "test" {
  name             = "my-test-notification-configuration"
  enabled          = true
  destination_type = "generic"
  triggers         = ["change_request:created"]
  url              = "https://example.com"
  team_id          = tfe_team.test.id
}
