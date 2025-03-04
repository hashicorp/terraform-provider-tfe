---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_run_task_global_settings"
description: |-
  Manages Run tasks global settings.
---

# tfe_organization_run_task_global_settings

[Run tasks](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks) allow HCP Terraform to interact with external systems at specific points in the HCP Terraform run lifecycle. Run tasks are reusable configurations that you can attach to any workspace in an organization.

The tfe_organization_run_task_global_settings resource creates, updates and destroys the [global settings](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks#global-run-tasks) for an [Organization Run task](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks#creating-a-run-task). Your organization must have the `global-run-task` [entitlement](https://developer.hashicorp.com/terraform/cloud-docs/api-docs#feature-entitlements) to use global run tasks.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization_run_task_global_settings" "example" {
  task_id = tfe_organization_run_task.example.id
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
```

## Argument Reference

The following arguments are supported:

* `enabled` - (Optional) Whether the run task will be applied globally.
* `enforcement_level` - (Required) The enforcement level of the global task. Valid values are `advisory` and `mandatory`.
* `stages` - (Required) The stages to run the task in. Valid values are one or more of `pre_plan`, `post_plan`, `pre_apply` and `post apply`.
* `task_id` - (Required) The id of the Run task which will have the global settings applied.

## Attributes Reference

* `id` - The ID of the global settings.

## Import

Run task global settings can be imported; use `<ORGANIZATION NAME>/<TASK NAME>` as the
import ID. For example:

```shell
terraform import tfe_organization_run_task_global_settings.test my-org-name/task-name
```
