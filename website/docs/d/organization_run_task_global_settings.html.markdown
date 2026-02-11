---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_run_task_global_settings"
description: |-
  Get information on a Run task's global settings.
---

# Data Source: tfe_organization_run_task_global_settings

[Run tasks](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks) allow HCP Terraform to interact with external systems at specific points in the HCP Terraform run lifecycle. Run tasks are reusable configurations that you can attach to any workspace in an organization.

The tfe_organization_run_task_global_settings resource creates, updates and destroys the [global settings](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks#global-run-tasks) for an [Organization Run task](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks#creating-a-run-task). Your organization must have the `global-run-task` [entitlement](https://developer.hashicorp.com/terraform/cloud-docs/api-docs#feature-entitlements) to use global run tasks.

## Example Usage

```hcl
data "tfe_organization_run_task_global_settings" "example" {
  task_id = "task-abc123"
}
```

## Argument Reference

The following arguments are supported:

* `task_id` - (Required) The id of the Run task with the global settings.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `enabled` - Whether the run task will be applied globally.
* `enforcement_level` - The enforcement level of the global task. Valid values are `advisory` and `mandatory`.
* `stages` - The stages to run the task in. Valid values are one or more of `pre_plan`, `post_plan`, `pre_apply` and `post apply`.
