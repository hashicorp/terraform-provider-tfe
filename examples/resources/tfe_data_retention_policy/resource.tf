# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Creating a data retention policy for a workspace

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@example.com"
}

resource "tfe_workspace" "test-workspace" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_data_retention_policy" "foobar" {
  workspace_id = tfe_workspace.test-workspace.id

  delete_older_than {
    days = 42
  }
}
