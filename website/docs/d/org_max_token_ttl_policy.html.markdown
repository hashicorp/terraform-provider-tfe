---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_org_max_token_ttl_policy"
description: |-
  Get information on an organization's maximum token TTL policy.
---

# Data Source: tfe_org_max_token_ttl_policy

Use this data source to retrieve information about an organization's maximum time-to-live (TTL) policy for API tokens. This policy defines the maximum lifespan for organization, team, audit trail, and user tokens.

## Example Usage

```hcl
data "tfe_org_max_token_ttl_policy" "example" {
  organization = "my-org-name"
}

output "org_token_ttl_ms" {
  value = data.tfe_org_max_token_ttl_policy.example.org_token_max_ttl_ms
}

output "policy_enabled" {
  value = data.tfe_org_max_token_ttl_policy.example.enabled
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `enabled` - Indicates whether the maximum TTL token policy is enabled (`true`) or disabled (`false`) for the organization.
* `org_token_max_ttl_ms` - Maximum lifespan allowed for organization tokens in milliseconds.
* `team_token_max_ttl_ms` - Maximum lifespan allowed for team tokens in milliseconds.
* `audit_trail_token_max_ttl_ms` - Maximum lifespan allowed for audit trail tokens in milliseconds.
* `user_token_max_ttl_ms` - Maximum lifespan allowed for user tokens in milliseconds.

## Notes

* When the policy is disabled, all TTL values default to `63072000000` milliseconds (2 years).
* The data source fetches the current policy configuration from the database via the TFE API.
* If no policies have been configured for the organization, the data source will return default values with `enabled = false`.
* TTL values are returned in milliseconds to preserve the exact values from the database without conversion ambiguity.