---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_run_task"
description: |-
  Manages Workspace Run tasks.
---

# tfe_workspace_run_task

[Run tasks](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks) allow Terraform Cloud to interact with external systems at specific points in the Terraform Cloud run lifecycle. Run tasks are reusable configurations that you can attach to any workspace in an organization.

The tfe_workspace_run_task resource associates, updates and removes [Workspace Run tasks](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks#associating-run-tasks-with-a-workspace).

## Example Usage

Basic usage:

```hcl
resource "tfe_workspace_run_task" "example" {
  workspace_id      = resource.tfe_workspace.example.id
  task_id           = resource.tfe_organization_run_task.example.id
  enforcement_level = "advisory"
}
```

## Argument Reference

The following arguments are supported:

* `enforcementLevel` - (Required) The enforcement level of the task. Valid values are `advisory` and `mandatory`.
* `taskId` - (Required) The id of the Run task to associate to the Workspace.
* `workspaceId` - (Required) The id of the workspace to associate the Run task to.
* `stage` - (Optional) The stage to run the task in. Valid values are `prePlan`, `postPlan`, and `preApply`.

## Attributes Reference

* `id` - The ID of the Workspace Run task.

## Import

Run tasks can be imported; use `<ORGANIZATION>/<WORKSPACE NAME>/<TASK NAME>` as the
import ID. For example:

```shell
terraform import tfe_workspace_run_task.test my-org-name/workspace/task-name
```

<!-- cache-key: cdktf-0.17.0-pre.15 input-c6eef700127257dfb3fc46f3f9dc706d6e4c14877b176dfe4e276fe83d59d5db -->