---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_agent_pool"
description: |-
  Manages agent pools
---

# tfe_agent_pool

An agent pool represents a group of agents, often related to one another by sharing a common
network segment or purpose. A workspace may be configured to use one of the organization's agent
pools to run remote operations with isolated, private, or on-premises infrastructure.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "test-agent-pool" {
  name         = "my-agent-pool-name"
  organization = tfe_organization.test-organization.name
  organization_scoped = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the agent pool.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `organization_scoped` - (Optional) Whether or not the agent pool is scoped to all workspaces in the organization. Defaults to `true`.

## Attributes Reference

* `id` - The ID of the agent pool.
* `name` - The name of agent pool.
* `organization` - The name of the organization associated with the agent pool.

## Import

Agent pools can be imported; use `<AGENT POOL ID>` or `<ORGANIZATION NAME>/<AGENT POOL NAME>` as the import ID. For example:

```shell
terraform import tfe_agent_pool.test apool-rW0KoLSlnuNb5adB
```

```shell
terraform import tfe_workspace.test my-org-name/my-agent-pool-name
```
