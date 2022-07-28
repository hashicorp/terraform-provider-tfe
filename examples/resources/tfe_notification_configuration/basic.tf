resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.id
}

resource "tfe_notification_configuration" "test" {
  name             = "my-test-notification-configuration"
  enabled          = true
  destination_type = "generic"
  triggers         = ["run:created", "run:planning", "run:errored"]
  url              = "https://example.com"
  workspace_id     = tfe_workspace.test.id
}