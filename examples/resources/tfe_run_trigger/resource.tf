# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_workspace" "test-sourceable" {
  name         = "my-sourceable-workspace-name"
  organization = tfe_organization.example.name
}

resource "tfe_run_trigger" "test" {
  workspace_id  = tfe_workspace.example.id
  sourceable_id = tfe_workspace.test-sourceable.id
}
