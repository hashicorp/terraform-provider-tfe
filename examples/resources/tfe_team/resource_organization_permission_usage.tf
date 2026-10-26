# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Organization Permission usage

resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
  organization_access {
    manage_vcs_settings = true
  }
}
