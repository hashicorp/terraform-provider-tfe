# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_policy_set" "test" {
  name         = "my-policy-set-name"
  organization = "my-org-name"
}
