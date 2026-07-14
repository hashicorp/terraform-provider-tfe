# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}
