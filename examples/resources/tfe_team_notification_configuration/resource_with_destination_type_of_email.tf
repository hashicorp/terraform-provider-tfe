resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = tfe_organization.test.id
}

data "tfe_organization_membership" "test" {
  organization = tfe_organization.test.name
  email        = "example@example.com"
}

resource "tfe_team_organization_member" "test" {
  team_id                    = tfe_team.test.id
  organization_membership_id = data.tfe_organization_membership.test.id
}

resource "tfe_team_notification_configuration" "test" {
  name             = "my-test-email-notification-configuration"
  enabled          = true
  destination_type = "email"
  email_user_ids   = [data.tfe_organization_membership.test.user_id]
  triggers         = ["change_request:created"]
  team_id          = tfe_team.test.id
}
