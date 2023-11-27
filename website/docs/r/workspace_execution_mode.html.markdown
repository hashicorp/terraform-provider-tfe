---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_execution_mode"
description: |-
  Manage the exeuction mode on a workspace
---

# tfe_workspace_execution_mode

Manages the exeuction mode on a workspace.

~> **Note:** This resource exists in order to resolve a mutual dependency issue that occurs when a workspace tries to access an agent pool. The [tfe_agent_pool_allowed_workspaces](https://registry.terraform.io/providers/hashicorp/tfe/latest/docs/resources/agent_pool_allowed_workspaces) resource is still required to, first, grant a workspace permission to access an agent pool when **organization_scoped** is set to false. This resources attaches the agent pool to the workspace, after permission has been granted.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test-workspace" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.name
}

resource "tfe_agent_pool" "test-agent-pool" {
  name                = "my-agent-pool-name"
  organization        = tfe_organization.test.name
  organization_scoped = false
}

resource "tfe_agent_pool_allowed_workspaces" "test-allowed-workspaces" {
  agent_pool_id         = tfe_agent_pool.test.id
  allowed_workspace_ids = [tfe_workspace.test-workspace.id]
}

resource "tfe_workspace_execution_mode" "test" {
  execution_mode = "agent"
  workspace_id = tfe_workspace.test-workspace.id
  agent_pool_id = tfe_agent_pool.test-workspace.id
}
```

## Argument Reference

The following arguments are supported:

* `execution_mode` - (Required) execution mode being set on the managed workspace. Accepted values are: `remote` (default), `local`, and `agent`. [Learn more about Execution Mode here](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode).
* `workspace_id` - (Required) ID of the workspace being managed.
* `agent_pool_id` - (Optional) ID of the agent pool being assigned to the managed workspace. Requires `execution_mode` to be set to `agent`. This value _must not_ be provided if `execution_mode` is set to any other mode, or if `operations` is provided.

## Attributes Reference

In addition to the arguments above, the following attribute is also exported:

* `id` - The workspace ID.

## Import

Workspaces can be imported; use `<WORKSPACE ID>` or `<ORGANIZATION NAME>/<WORKSPACE NAME>` as the import ID. For example:

```shell
terraform import tfe_workspace_execution_mode.test ws-CH5in3chf8RJjrVd
```

```shell
terraform import tfe_workspace_execution_mode.test my-org-name/my-wkspace-name
```