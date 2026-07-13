# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = "my-org-name"
}
