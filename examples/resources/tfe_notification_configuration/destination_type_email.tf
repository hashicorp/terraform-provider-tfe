resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.id
}

resource "tfe_organization_membership" "test" {
  organization = "my-org-name"
  email        = "test.member@company.com"
}

resource "tfe_notification_configuration" "test" {
  name             = "my-test-email-notification-configuration"
  enabled          = true
  destination_type = "email"
  email_user_ids   = [tfe_organization_membership.test.user_id]
  triggers         = ["run:created", "run:planning", "run:errored"]
  workspace_id     = tfe_workspace.test.id
}