---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_ip_allowlist"
description: |-
  Get information on an IP allowlist (CIDR range list).
---

# Data Source: tfe_ip_allowlist

Use this data source to get information about an IP allowlist (referred to as a
CIDR range list in the HCP Terraform API) in an organization.

## Example Usage

```hcl
data "tfe_ip_allowlist" "example" {
  name         = "corporate-network"
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the IP allowlist.
* `organization` - (Optional) Name of the organization. If omitted, organization
  must be defined in the provider config.

## Attributes Reference

* `id` - The ID of the IP allowlist.
* `description` - A description for the IP allowlist.
* `enforcement_scope` - Where the IP allowlist is enforced. One of
  `organization`, `all_agent_pools`, or `selected_agent_pools`.
* `agent_pool_ids` - The IDs of the agent pools the IP allowlist applies to.
* `cidr_range` - The set of CIDR ranges that belong to the allowlist. Each entry
  contains:
  * `range` - The IPv4 CIDR range.
  * `description` - A description for the CIDR range.
  * `enabled` - Whether the CIDR range is enforced.
