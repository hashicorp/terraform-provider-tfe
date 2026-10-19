# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Tracking a local directory to upload the Sentinel configuration and policies

data "tfe_slug" "test" {
  source_path = "policies/my-policy-set"
}

resource "tfe_policy_set" "test" {
  name         = "my-policy-set"
  organization = "my-org-name"

  // reference the tfe_slug data source.
  slug = data.tfe_slug.test
}
