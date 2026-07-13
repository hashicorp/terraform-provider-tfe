# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.name
}

resource "tfe_variable_set" "test" {
  name         = "Test Varset"
  description  = "Some description."
  organization = tfe_organization.test.name
}

resource "tfe_workspace_variable_set" "test" {
  variable_set_id = tfe_variable_set.test.id
  workspace_id    = tfe_workspace.test.id
}
