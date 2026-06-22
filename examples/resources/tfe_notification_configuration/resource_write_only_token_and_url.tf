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
  destination_type = "generic"
  token_wo         = "my-secret-token"
  url_wo           = "https://example.com"
  workspace_id     = tfe_workspace.test.id
}