---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_ip_allowlist"
description: |-
  Manages IP allowlists (CIDR range lists) in an organization.
---

# tfe_ip_allowlist

Creates, updates and destroys IP allowlists (referred to as CIDR range lists in
the HCP Terraform API). An IP allowlist restricts which client IP addresses may
access HCP Terraform for an organization or for specific agent pools.

~> **NOTE:** Managing IP allowlists requires organization owner access. API
tokens with limited scopes (such as team or organization tokens) cannot manage
IP allowlists.

~> **NOTE:** Only IPv4 CIDR ranges are supported.

## Example Usage

Organization-wide allowlist:

```hcl
resource "tfe_ip_allowlist" "example" {
  organization      = "my-org-name"
  name              = "corporate-network"
  description       = "Allowlist for the corporate network"
  enforcement_scope = "organization"

  cidr_range = [
    {
      range       = "10.0.0.0/16"
      description = "Corporate LAN"
    },
    {
      range       = "192.168.1.0/24"
      description = "VPN"
      enabled     = false
    },
  ]
}
```

Allowlist scoped to selected agent pools:

```hcl
resource "tfe_agent_pool" "example" {
  name         = "my-agent-pool"
  organization = "my-org-name"
}

resource "tfe_ip_allowlist" "example" {
  organization      = "my-org-name"
  name              = "agent-pool-allowlist"
  enforcement_scope = "selected_agent_pools"
  agent_pool_ids    = [tfe_agent_pool.example.id]

  cidr_range = [
    {
      range = "203.0.113.0/24"
    },
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the IP allowlist. Must be unique within the
  organization.
* `enforcement_scope` - (Required) Where the IP allowlist is enforced. Must be
  one of `organization`, `all_agent_pools`, or `selected_agent_pools`. Only one
  `organization`-scoped allowlist may exist per organization.
* `cidr_range` - (Required) A set of one or more CIDR ranges that belong to the
  allowlist. Each block supports the fields documented under
  [CIDR ranges](#cidr-ranges).
* `description` - (Optional) A description for the IP allowlist.
* `organization` - (Optional) Name of the organization. If omitted, organization
  must be defined in the provider config.
* `agent_pool_ids` - (Optional) The IDs of the agent pools the IP allowlist
  applies to. Only valid when `enforcement_scope` is `selected_agent_pools`.

### CIDR ranges

The `cidr_range` set must contain at least one entry. Each entry supports the
following:

* `range` - (Required) An IPv4 CIDR range, e.g. `10.0.0.0/24`. The CIDR value is
  the identity of the range: changing it removes the old range and creates a new
  one, whereas changing `description` or `enabled` updates the existing range
  in place.
* `description` - (Optional) A description for the CIDR range. If omitted, the
  range has no description (it is not defaulted to an empty string).
* `enabled` - (Optional) Whether the CIDR range is enforced. Defaults to `true`.

## Attributes Reference

* `id` - The ID of the IP allowlist.

## Import

IP allowlists can be imported by their ID, e.g.

```shell
terraform import tfe_ip_allowlist.example crl-EXAMPLE1234
```
