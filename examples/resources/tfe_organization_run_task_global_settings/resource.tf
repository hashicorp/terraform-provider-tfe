# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_organization_run_task_global_settings" "example" {
  task_id           = tfe_organization_run_task.example.id
  enabled           = true
  enforcement_level = "advisory"
  stages            = ["pre_plan", "post_plan"]
}

resource "tfe_organization_run_task" "example" {
  organization = "org-name"
  url          = "https://external.service.com"
  name         = "task-name"
  enabled      = true
  description  = "An example task"
}
