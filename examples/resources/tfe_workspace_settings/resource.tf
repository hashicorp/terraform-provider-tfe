# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_workspace_settings" "test-settings" {
  workspace_id   = tfe_workspace.example.id
  execution_mode = "local"
}
