# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_slug" "test" {
  source_path = "policies/my-policy-set"
}

resource "tfe_policy_set" "test" {
  name         = "my-policy-set"
  organization = "my-org-name"

  // reference the tfe_slug data source.
  slug = data.tfe_slug.test
}
