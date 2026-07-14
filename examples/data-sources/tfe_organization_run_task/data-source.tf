# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_organization_run_task" "example" {
  name         = "task-name"
  organization = "my-org-name"
}
