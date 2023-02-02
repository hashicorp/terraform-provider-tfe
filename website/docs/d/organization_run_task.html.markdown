---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_run_task"
description: |-
  Get information on a Run task.
---

# Data Source: tfe_organization_run_task

[Run tasks](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks) allow Terraform Cloud to interact with external systems at specific points in the Terraform Cloud run lifecycle. Run tasks are reusable configurations that you can attach to any workspace in an organization.

Use this data source to get information about an [Organization Run tasks](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks#creating-a-run-task).

## Example Usage

```hcl
data "tfe_organization_run_task" "example" {
  name         = "task-name"
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the Run task.
* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `category` - The type of task.
* `description` - A short description of the the task.
* `enabled` - Whether the task will be run.
* `id` - The ID of the task.
* `url` - URL to send a task payload.
