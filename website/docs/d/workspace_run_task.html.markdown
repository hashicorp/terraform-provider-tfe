---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_task"
description: |-
  Get information on a Workspace Run task.
---

# Data Source: tfe_workspace_task

[Run tasks](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks) allow Terraform Cloud to interact with external systems at specific points in the Terraform Cloud run lifecycle. Run tasks are reusable configurations that you can attach to any workspace in an organization.

Use this data source to get information about a [Workspace Run tasks](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks#associating-run-tasks-with-a-workspace).

## Example Usage

```hcl
data "tfe_workspace_run_task" "foobar" {
  workspace_id      = "ws-abc123"
  task_id           = "task-def456"
}
```

## Argument Reference

The following arguments are supported:

* `task_id` - (Required) The id of the run task.
* `workspace_id` - (Required) The id of the workspace.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `enforcement_level` - The enforcement level of the task.
* `id` - The ID of the Workspace Run task.
* `stage` - Which stage the task will run in.
