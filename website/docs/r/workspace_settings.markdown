---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_setting"
description: |-
  Manages workspace settings.
---

# tfe_workspace_settings

~> **NOTE:** Manages or reads execution mode and agent pool settings for a workspace. This also interacts with the organization's default values for several settings, which can be managed with [tfe_organization_default_settings](organization_default_settings.html). If other resources need to identify whether a setting is a default or an explicit value set for the workspace, you can refer to the read-only `overwrites` argument.

~> **NOTE:** This resource manages values that can alternatively be managed by the  `tfe_workspace` resource. You should not attempt to manage the same property on both resources which could cause a permanent drift. Example properties available on both resources: `description`, `tags`, `auto_apply`, etc.

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
  agent_pool_id  = tfe_agent_pool_allowed_workspaces.test.agent_pool_id
  execution_mode = "agent"
}
```

Using `remote_state_consumer_ids`:

```hcl
resource "tfe_workspace" "test" {
  for_each                  = toset(["qa","production"])
  name                      = "${each.value}-test"
}

resource "tfe_workspace_settings" "test-settings" {
  for_each                  = toset(["qa","production"])
  workspace_id              = tfe_workspace.test[each.value].id
  global_remote_state       = false
  remote_state_consumer_ids = toset(compact([each.value == "production" ? tfe_workspace.test["qa"].id : ""]))
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

This resource can be used to self manage a workspace created from `terraform init` and a cloud block:

```hcl
terraform {
  cloud {
    organization = "foo"
    workspaces {
      name = "self-managed"
    }
  }
}

# workspace is created in CI during `init`
data "tfe_workspace" "self" {
  name         = split("/", var.TFC_WORKSPACE_SLUG)[1]
  organization = split("/", var.TFC_WORKSPACE_SLUG)[0]
}

# settings and notification for workspace are applied 
resource "tfe_workspace_settings" "self" {
  workspace_id        = data.tfe_workspace.self.id
  assessments_enabled = true
  tags = {
    prod = "true"
  }
}
```

## Argument Reference

The following arguments are supported:

* `workspace_id` - (Required) ID of the workspace.
* `agent_pool_id` - (Optional) The ID of an agent pool to assign to the workspace. Requires `execution_mode`
  to be set to `agent`. This value _must not_ be provided if `execution_mode` is set to any other value.
* `execution_mode` - (Optional) Which [execution mode](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode)
  to use. Using HCP Terraform, valid values are `remote`, `local` or `agent`. When set to `local`, the workspace will be used for state storage only. **Important:** If you omit this attribute, the resource configures the workspace to use your organization's default execution mode (which in turn defaults to `remote`), removing any explicit value that might have previously been set for the workspace.
* `global_remote_state` - (Optional) Whether the workspace allows all workspaces in the organization to access its state data during runs. If false, then only specifically approved workspaces can access its state (`remote_state_consumer_ids`). By default, HashiCorp recommends you do not allow other workspaces to access their state. We recommend that you follow the principle of least privilege and only enable state access between workspaces that specifically need information from each other.
* `remote_state_consumer_ids` - (Optional) The set of workspace IDs set as explicit remote state consumers for the given workspace. To set this attribute, global_remote_state must be false.
* `auto_apply` - (Optional) Whether to automatically apply changes when a Terraform plan is successful. Defaults to `false`.
* `assessments_enabled` - (Optional) Whether to regularly run health assessments such as drift detection on the workspace. Defaults to `false`.
* `description` - (Optional) A description for the workspace.
* `tags` - (Optional) A map of key value tags for this workspace.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The workspace ID.
* `overwrites` - Can be used to check whether a setting is currently inheriting its value from another resource.
  - `execution_mode` - Set to `true` if the execution mode of the workspace is being determined by the setting on the workspace itself. It will be `false` if the execution mode is inherited from another resource (e.g. the organization's default execution mode)
  - `agent_pool` - Set to `true` if the agent pool of the workspace is being determined by the setting on the workspace itself. It will be `false` if the agent pool is inherited from another resource (e.g. the organization's default agent pool)
* `effective_tags` - A map of key value tags for this workspace, including any tags inherited from the parent project.

## Import

Workspaces can be imported; use `<WORKSPACE ID>` or `<ORGANIZATION NAME>/<WORKSPACE NAME>` as the
import ID. For example:

```shell
terraform import tfe_workspace_settings.test ws-CH5in3chf8RJjrVd
```

```shell
terraform import tfe_workspace_settings.test my-org-name/my-wkspace-name
```
