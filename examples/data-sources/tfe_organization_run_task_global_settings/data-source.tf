# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_organization_run_task_global_settings" "example" {
  task_id = "task-abc123"
}
