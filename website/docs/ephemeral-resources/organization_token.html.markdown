---
layout: "tfe"
page_title: "Terraform Enterprise: Ephemeral: tfe_organization_token"
description: |-
  Generates a new organization token that is guaranteed not to be written to
  state.
---

# Ephemeral: tfe_organization_token

Terraform ephemeral resource for managing a TFE organization token. This
resource is used to generate a new organization token that is guaranteed not to
be written to state. Since organization tokens are singleton resources, using this ephemeral resource will replace any existing organization token.

~> **NOTE:** Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/v1.10.x/resources/ephemeral).

## Example Usage

### Generate a new organization token:

This will invalidate any existing organization token.

```hcl
resource "tfe_organization_token" "example" {
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are required:

* `organization` - (Required) Name of the organization. If omitted, organization must be defined in the provider config.

The following arguments are optional:

* `expired_at` - (Optional) The token's expiration date. The expiration date must be a date/time string in RFC3339 
format (e.g., "2024-12-31T23:59:59Z"). If no expiration date is supplied, the expiration date will default to null and 
never expire.

This ephemeral resource exports the following attributes in addition to the arguments above:

* `token` - The generated token. This value is sensitive and will not be stored
  in state.
