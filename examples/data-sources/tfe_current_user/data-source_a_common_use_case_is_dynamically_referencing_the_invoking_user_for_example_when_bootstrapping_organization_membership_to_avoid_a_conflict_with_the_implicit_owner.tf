# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_current_user" "current" {}

resource "tfe_organization_membership" "owner" {
  organization = "my-org"
  email        = data.tfe_current_user.current.email
}
