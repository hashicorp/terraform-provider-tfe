# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# With write-only token and URL (auto-managed, recommended)

resource "tfe_notification_configuration" "test" {
  name             = "my-test-notification-configuration"
  destination_type = "generic"
  token_wo         = "my-secret-token"
  url_wo           = "https://example.com"
  workspace_id     = tfe_workspace.example.id
}
