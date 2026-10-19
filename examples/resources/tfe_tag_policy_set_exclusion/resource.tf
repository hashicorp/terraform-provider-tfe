# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@example.com"
}

resource "tfe_workspace" "test" {
  name         = "tagged-workspace"
  organization = tfe_organization.test.name

  tag_names = ["env", "others", "staging"]
}

resource "tfe_policy_set" "test" {
  name         = "excluded-policy-set"
  description  = "Some description."
  organization = tfe_organization.test.name
  global       = true
}

resource "tfe_tag_policy_set_exclusion" "test" {
  policy_set_id = tfe_policy_set.test.id
  key           = "env"
  value         = "staging"
}
