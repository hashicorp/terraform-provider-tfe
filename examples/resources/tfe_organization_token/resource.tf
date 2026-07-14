# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_organization_token" "test" {
  organization = "my-org-name"
}
