# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_workspace" "test" {
  organization = "my-org-name"
  name         = "my-workspace"
}

resource "tfe_workspace_hyok_enabled" "test" {
  workspace_id = tfe_workspace.test.id
}
