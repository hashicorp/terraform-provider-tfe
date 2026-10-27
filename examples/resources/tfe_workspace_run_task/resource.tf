# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_workspace" "ws" {
  name         = "example-workspace"
  organization = "my-org-name"
}

resource "tfe_organization_run_task" "example" {
  organization = "org-name"
  url          = "https://external.service.com"
  name         = "task-name"
  enabled      = true
  description  = "An example task"
}

resource "tfe_workspace_run_task" "example" {
  workspace_id      = resource.tfe_workspace.ws.id
  task_id           = resource.tfe_organization_run_task.example.id
  enforcement_level = "advisory"
  stages            = ["pre_plan"]
}
