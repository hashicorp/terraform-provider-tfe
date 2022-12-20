---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_agent_pool"
description: |-
  Get information on an agent pool.
---

# Data Source: tfe_agent_pool

Use this data source to get information about an agent pool.

~> **NOTE:** This data source requires using the provider with Terraform Cloud and a Terraform Cloud 
for Business account. 
[Learn more about Terraform Cloud pricing here](https://www.hashicorp.com/products/terraform/pricing).

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