# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_workspace" "example" {
  name         = "example-workspace"
  organization = "my-organization"
}

resource "tfe_organization_run_task" "example" {
  organization = "org-name"
  url          = "https://external.service.com"
  name         = "task-name"
  enabled      = true
  description  = "An example task"
}

resource "tfe_workspace_run_task" "example" {
  workspace_id      = resource.tfe_workspace.example.id
  task_id           = resource.tfe_organization_run_task.example.id
  enforcement_level = "advisory"
  stages            = ["pre_plan"]
}
