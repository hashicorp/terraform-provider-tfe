---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_default_execution_mode"
description: |-
  Sets the default execution mode of an organization
---

# tfe_organization_default_execution_mode

Sets the default execution mode of an organization. This default execution mode will be used as the default execution mode for all workspaces in the organization.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "my_agents" {
  name = "agent_smiths"
  organization = tfe_organization.test.name
}

resource "tfe_organization_default_execution_mode" "org_default" {
  organization = tfe_organization.test.name
  default_execution_mode = "agent"
  default_agent_pool_id = tfe_agent_pool.my_agents.id
}
```

## Argument Reference

The following arguments are supported:

* `default_execution_mode` - (Optional) Which [execution mode](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode)
  to use as the default for all workspaces in the organization. Valid values are `remote`, `local` or`agent`.
* `agent_pool_id` - (Optional) The ID of an agent pool to assign to the workspace. Requires `default_execution_mode`
  to be set to `agent`. This value _must not_ be provided if `default_execution_mode` is set to any other value or if `operations` is
  provided.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.


## Import

This resource does not manage the creation of an organization and there is no need to import it.