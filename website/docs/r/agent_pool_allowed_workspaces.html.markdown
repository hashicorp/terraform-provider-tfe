---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_agent_pool_allowed_workspaces"
description: |-
  Manages allowed workspaces on agent pools
---

# tfe_agent_pool_allowed_workspaces

Adds and removes allowed workspaces on an agent pool.

~> **NOTE:** This resource requires using the provider with HCP Terraform and a HCP Terraform
for Business account.
[Learn more about HCP Terraform pricing here](https://www.hashicorp.com/products/terraform/pricing).

## Example Usage

In this example, the agent pool and workspace are connected through other resources that manage the agent pool permissions as well as the workspace execution mode. Notice that the `tfe_workspace_settings` uses the agent pool reference found in `tfe_agent_pool_allowed_workspaces` in order to create the permission to use the agent pool before assigning it.

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

// Ensure workspace and agent pool are create first
resource "tfe_workspace" "test-workspace" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_agent_pool" "test-agent-pool" {
  name                = "my-agent-pool-name"
  organization        = tfe_organization.test-organization.name
  organization_scoped = false
}

// Ensure permissions are assigned second
resource "tfe_agent_pool_allowed_workspaces" "allowed" {
  agent_pool_id         = tfe_agent_pool.test-agent-pool.id
  allowed_workspace_ids = [tfe_workspace.test-workspace.id]
}

// Lastly, ensure the workspace agent execution is assigned last by
// referencing allowed_workspaces
resource "tfe_workspace_settings" "test-workspace-settings" {
  workspace_id   = tfe_workspace.test-workspace.id
  execution_mode = "agent"
  agent_pool_id  = tfe_agent_pool_allowed_workspaces.allowed.id
}
```

## Argument Reference

The following arguments are supported:

* `agent_pool_id` - (Required) The ID of the agent pool.
* `allowed_workspace_ids` - (Required) IDs of workspaces to be added as allowed workspaces on the agent pool.


## Import

A resource can be imported; use `<AGENT POOL ID>` as the import ID. For example:

```shell
terraform import tfe_agent_pool_allowed_workspaces.foobar apool-rW0KoLSlnuNb5adB
```
