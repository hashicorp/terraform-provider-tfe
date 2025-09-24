---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_agent_pool_excluded_workspaces"
description: |-
  Manages excluded workspaces on agent pools
---

# tfe_agent_pool_excluded_workspaces

Adds and removes excluded workspaces on an agent pool.

~> **NOTE:** This resource requires using the provider with HCP Terraform and a HCP Terraform
for Business account.
[Learn more about HCP Terraform pricing here](https://www.hashicorp.com/products/terraform/pricing).

## Example Usage

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
resource "tfe_agent_pool_excluded_workspaces" "excluded" {
  agent_pool_id          = tfe_agent_pool.test-agent-pool.id
  excluded_workspace_ids = [tfe_workspace.test-workspace.id]
}
```

## Argument Reference

The following arguments are supported:

* `agent_pool_id` - (Required) The ID of the agent pool.
* `excluded_workspace_ids` - (Required) IDs of workspaces to be added as excluded workspaces on the agent pool.


## Import

A resource can be imported; use `<AGENT POOL ID>` as the import ID. For example:

```shell
terraform import tfe_agent_pool_excluded_workspaces.foobar apool-rW0KoLSlnuNb5adB
```
