# (Terraform Enterprise only) With destination_type of email, using email_addresses list and email_users

resource "tfe_organization_membership" "test" {
  organization = "my-org-name"
  email        = "test.member@example.com"
}

resource "tfe_notification_configuration" "test" {
  name             = "my-test-email-notification-configuration"
  enabled          = true
  destination_type = "email"
  email_user_ids   = [tfe_organization_membership.test.user_id]
  email_addresses  = ["user1@example.com", "user2@example.com", "user3@example.com"]
  triggers         = ["run:created", "run:planning", "run:errored"]
  workspace_id     = tfe_workspace.example.id
}
