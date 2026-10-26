# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_organization_membership" "test" {
  organization = "my-org-name"
  email        = "user@example.com"
}
