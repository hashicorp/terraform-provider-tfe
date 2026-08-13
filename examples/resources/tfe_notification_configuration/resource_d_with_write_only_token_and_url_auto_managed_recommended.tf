# With write-only token and URL (auto-managed, recommended)

resource "tfe_notification_configuration" "test" {
  name             = "my-test-notification-configuration"
  destination_type = "generic"
  token_wo         = "my-secret-token"
  url_wo           = "https://example.com"
  workspace_id     = tfe_workspace.example.id
}
