# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

// create data retention policy the organization
resource "tfe_data_retention_policy" "foo" {
  organization = tfe_organization.test-organization.name

  delete_older_than {
    days = 1138
  }
}

resource "tfe_workspace" "test-workspace" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.name
}

// create a policy that prevents automatic deletion of data in the test-workspace
resource "tfe_data_retention_policy" "bar" {
  workspace_id = tfe_workspace.test-workspace.id

  dont_delete {}
}
