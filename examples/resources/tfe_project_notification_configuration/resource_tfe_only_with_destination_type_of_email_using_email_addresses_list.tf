# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_project" "test" {
  name         = "my-project-name"
  organization = tfe_organization.test.id
}

resource "tfe_organization_membership" "test" {
  organization = "my-org-name"
  email        = "test.member@company.com"
}

resource "tfe_project_notification_configuration" "test" {
  name             = "my-test-email-notification-configuration"
  enabled          = true
  destination_type = "email"
  email_user_ids   = [tfe_organization_membership.test.user_id]
  email_addresses  = ["user1@company.com", "user2@company.com", "user3@company.com"]
  triggers         = ["run:created", "run:planning", "run:errored"]
  project_id       = tfe_project.test.id
}
