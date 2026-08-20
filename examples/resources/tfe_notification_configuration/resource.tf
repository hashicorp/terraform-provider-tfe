# Basic usage

resource "tfe_notification_configuration" "test" {
  name             = "my-test-notification-configuration"
  enabled          = true
  destination_type = "generic"
  triggers         = ["run:created", "run:planning", "run:errored"]
  url_wo           = "https://example.com"
  workspace_id     = tfe_workspace.example.id
}
