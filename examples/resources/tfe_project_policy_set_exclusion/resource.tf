# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_policy_set" "test" {
  name         = "my-policy-set"
  description  = "Some description."
  organization = tfe_organization.example.name
  global       = true
}

resource "tfe_project_policy_set_exclusion" "test" {
  policy_set_id = tfe_policy_set.test.id
  project_id    = tfe_project.example.id
}
