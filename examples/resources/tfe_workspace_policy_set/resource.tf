# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_policy_set" "example" {
  name         = "my-policy-set"
  description  = "Example fixture policy set."
  organization = tfe_organization.example.name
}

resource "tfe_workspace_policy_set" "test" {
  policy_set_id = tfe_policy_set.example.id
  workspace_id  = tfe_workspace.example.id
}
