# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_workspace_run_task" "foobar" {
  workspace_id = "ws-abc123"
  task_id      = "task-def456"
}
