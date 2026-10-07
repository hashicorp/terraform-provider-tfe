# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_variable_set" "example" {
  name         = "my-variable-set"
  description  = "Example fixture variable set."
  organization = tfe_organization.example.name
}

resource "tfe_workspace_variable_set" "test" {
  variable_set_id = tfe_variable_set.example.id
  workspace_id    = tfe_workspace.example.id
}
