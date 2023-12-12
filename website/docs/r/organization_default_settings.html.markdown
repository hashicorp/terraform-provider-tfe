---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_default_settings
description: |-
  Sets the workspace defaults for an organization
---

# tfe_organization_default_settings

Primarily, this is used to set the default execution mode of an organization. This setting will be used as the default for all workspaces in the organization.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "my_agents" {
  name         = "agent_smiths"
  organization = tfe_organization.test.name
}

resource "tfe_organization_default_settings" "org_default" {
  organization           = tfe_organization.test.name
  default_execution_mode = "agent"
  default_agent_pool_id  = tfe_agent_pool.my_agents.id
}

resource "tfe_workspace" "my_workspace" {
  name       = "my-workspace"
  # Ensures this workspace will inherit the org defaults
  depends_on = [tfe_organization_default_settings.org_default]
}
```

## Argument Reference

The following arguments are supported:

* `default_execution_mode` - (Optional) Which [execution mode](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode)
  to use as the default for all workspaces in the organization. Valid values are `remote`, `local` or`agent`.
* `default_agent_pool_id` - (Optional) The ID of an agent pool to assign to the workspace. Requires `default_execution_mode` to be set to `agent`. This value _must not_ be provided if `default_execution_mode` is set to any other value.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.


## Import

Organization default execution mode can be imported; use `<ORGANIZATION NAME>` as the import ID. For example:

```shell
terraform import tfe_organization_default_execution_mode.test my-org-name
```
