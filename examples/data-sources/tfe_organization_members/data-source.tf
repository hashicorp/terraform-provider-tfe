# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_organization" "bar" {
  name  = "org-bar"
  email = "user@hashicorp.com"
}

data "tfe_organization_members" "foo" {
  organization = tfe_organization.bar.name
}
