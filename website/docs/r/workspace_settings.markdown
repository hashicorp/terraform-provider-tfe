---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_setting"
description: |-
  Manages workspace settings.
---

# tfe_workspace_settings

Manages or reads execution mode and agent pool settings for a workspace. If [tfe_organization_default_settings](organization_default_settings.html) are used, those settings may be read using a combination of the read-only `overwrites` argument and the setting itself.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_workspace_settings" "test-settings" {
  workspace_id   = tfe_workspace.test.id
  execution_mode = "local"
}
```

With `execution_mode` of `agent`:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "test-agent-pool" {
  name         = "my-agent-pool-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_agent_pool_allowed_workspaces" "test" {
  agent_pool_id         = tfe_agent_pool.test-agent-pool.id
  allowed_workspace_ids = [tfe_workspace.test.id]
}

resource "tfe_workspace" "test" {
  name           = "my-workspace-name"
  organization   = tfe_organization.test-organization.name
}

resource "tfe_workspace_settings" "test-settings" {
  workspace_id   = tfe_workspace.test.id
  agent_pool_id  = tfe_agent_pool.test-agent-pool.id
  execution_mode = "agent"
}
```

This resource may be used as a data source when no optional arguments are defined:

```hcl
data "tfe_workspace" "test" {
  name           = "my-workspace-name"
  organization   = "my-org-name"
}

resource "tfe_workspace_settings" "test" {
  workspace_id   = data.tfe_workspace.test.id
}

output "workspace-explicit-local-execution" {
  value = alltrue([
    tfe_workspace_settings.test.execution_mode == "local",
    tfe_workspace_settings.test.overwrites[0]["execution_mode"]
  ])
}
```

## Argument Reference

The following arguments are supported:

* `workspace_id` - (Required) ID of the workspace.
* `agent_pool_id` - (Optional) The ID of an agent pool to assign to the workspace. Requires `execution_mode`
  to be set to `agent`. This value _must not_ be provided if `execution_mode` is set to any other value.
* `execution_mode` - (Optional) Which [execution mode](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode)
  to use. Using Terraform Cloud, valid values are `remote`, `local` or `agent`. Defaults  your organization's default execution mode, or `remote` if no organization default is set. Using Terraform Enterprise, only `remote` and `local` execution modes are valid.  When set to `local`, the workspace will be used for state storage only.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The workspace ID.
* `overwrites` - Can be used to check whether a setting is currently inheriting its value from another resource.
  - `execution_mode` - Set to `true` if the execution mode of the workspace is being determined by the setting on the workspace itself. It will be `false` if the execution mode is inherited from another resource (e.g. the organization's default execution mode)
  - `agent_pool` - Set to `true` if the agent pool of the workspace is being determined by the setting on the workspace itself. It will be `false` if the agent pool is inherited from another resource (e.g. the organization's default agent pool)

## Import

Workspaces can be imported; use `<WORKSPACE ID>` or `<ORGANIZATION NAME>/<WORKSPACE NAME>` as the
import ID. For example:

```shell
terraform import tfe_workspace_settings.test ws-CH5in3chf8RJjrVd
```

```shell
terraform import tfe_workspace_settings.test my-org-name/my-wkspace-name
```
