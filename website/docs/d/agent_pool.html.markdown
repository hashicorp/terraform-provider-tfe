---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_agent_pool"
description: |-
  Get information on an agent pool.
---

# Data Source: tfe_agent_pool

Use this data source to get information about an agent pool.

## Example Usage

```hcl
data "tfe_agent_pool" "test" {
  name          = "my-agent-pool-name"
  organization  = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the agent pool.
* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The agent pool ID.
* `allowed_project_ids` - The set of project IDs that have permission to use the agent pool.
* `allowed_workspace_ids` - The set of workspace IDs that have permission to use the agent pool.
* `excluded_workspace_ids` - The set of workspace IDs that are excluded from the scope of the agent pool.
* `organization_scoped` - Whether or not the agent pool can be used by all workspaces in the organization.
