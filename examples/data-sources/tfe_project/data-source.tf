# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_project" "foo" {
  name         = "my-project-name"
  organization = "my-org-name"
}
