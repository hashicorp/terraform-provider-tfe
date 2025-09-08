---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_project_settings"
description: |-
    Manage Project Settings.
---

# tfe_project_settings

Use this resource to manage Project Settings.

Primarily, this resource allows setting default execution mode and agent pool for all workspaces within a project. When not specified, the organization defaults will be used.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"

  # this will end up being overwritten at the project level
  default_execution_mode = "remote"
}

resource "tfe_agent_pool" "my_agents" {
  name         = "my-agent-pool"
  organization = tfe_organization.test.name
}

resource "tfe_project" "my_project" {
  name         = "my-project"
  organization = tfe_organization.test.name
}

resource "tfe_project_settings" "my_project_settings" {
  project_id             = tfe_project.my_project.id

  # workspaces in this project will use agent execution mode by default,
  # and will use the specified agent pool.
  default_execution_mode = "agent"
  default_agent_pool_id  = tfe_agent_pool.my_agents.id
}
```

## Argument Reference

The following arguments are supported:
* `project_id` - (Required) The ID of the project to manage settings for.
* `default_execution_mode` - (Optional) Which [execution mode](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode)
  to use as the default for all workspaces in the project. Valid values are `remote`, `local` or `agent`.
* `default_agent_pool_id` - (Optional) The ID of an agent pool to assign to the workspace. Requires `default_execution_mode` to be set to `agent`. This value _must not_ be provided if `default_execution_mode` is set to any other value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:
* `overwrites` - Can be used to check whether a setting is currently inheriting its value from the organization.
  - `default_execution_mode` - Set to `true` if the default execution mode of the project is being determined by the setting on the project itself. It will be `false` if the execution mode is inherited from another resource (e.g. the organization's default execution mode)
  - `default_agent_pool_id` - Set to `true` if the default agent pool of the project is being determined by the setting on the project itself. It will be `false` if the agent pool is inherited from another resource (e.g. the organization's default agent pool)parent project.

## Import

Project settings  can be imported; use the `<PROJECT_ID>` as the import ID. For example:

```shell
terraform import tfe_project_settings.my_project_settings <PROJECT_ID>
```
