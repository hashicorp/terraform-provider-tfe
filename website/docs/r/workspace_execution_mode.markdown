---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_execution_mode"
description: |-
  Change the execution mode of a workspace to use a particular agent pool
---

# tfe_workspace_execution_mode

Changes the execution mode of a workspace to use a particular agent pool.

-> **Note:** This resource exists in order to resolve a mutual dependency between workspaces and agent pools.

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
  workspace_id = tfe_workspace.test-workspace.id
  agent_pool_id = tfe_agent_pool.test-workspace.id
}
```

## Argument Reference

The following arguments are supported:

* `agent_pool_id` - (Required) ID of the agent pool to execute workspace runs on.
* `workspace_id` - (Required) workspace ID to change the execution mode of

## Import

Excluded Workspace Policy Sets can be imported; use `<ORGANIZATION>/<WORKSPACE NAME>`. For example:

```shell
terraform import tfe_workspace_execution_mode.test 'my-org-name/workspace-name'
```
