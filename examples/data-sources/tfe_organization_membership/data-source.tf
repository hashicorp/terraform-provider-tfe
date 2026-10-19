# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Fetch by email

data "tfe_organization_membership" "test" {
  organization = "my-org-name"
  email        = "user@example.com"
}
