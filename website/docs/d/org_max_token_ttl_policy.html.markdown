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
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `org_token_max_ttl_ms` - Maximum lifespan allowed for organization tokens in milliseconds.
* `team_token_max_ttl_ms` - Maximum lifespan allowed for team tokens in milliseconds.
* `audit_trail_token_max_ttl_ms` - Maximum lifespan allowed for audit trail tokens in milliseconds.
* `user_token_max_ttl_ms` - Maximum lifespan allowed for user tokens in milliseconds.

## Notes

* To check if the maximum TTL policy feature is enabled for an organization, use the `max_ttl_enabled` attribute on the `tfe_organization` data source.
* The data source fetches the current policy configuration from the TFE API.
* If no policies have been configured for the organization, the data source will return default values (2 years for all token types).
* TTL values are returned in milliseconds to preserve the exact values from the API without conversion ambiguity.