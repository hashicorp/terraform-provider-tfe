---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_audit_trail_token"
description: |-
  Generates a new audit trail token in organization, replacing any existing token.
---

# tfe_audit_trail_token

Generates a new audit trail token in organization, replacing any existing token.

Note that only organizations that have the [audit-logging entitlement](https://developer.hashicorp.com/terraform/cloud-docs/api-docs#audit-logging) may create audit trail tokens.

## Example Usage

Basic usage:

```hcl
resource "tfe_audit_trail_token" "test" {
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `force_regenerate` - (Optional) If set to `true`, a new token will be
  generated even if a token already exists. This will invalidate the existing
  token!
* `expired_at` - (Optional) The token's expiration date. The expiration date must be a date/time string in RFC3339
format (e.g., "2024-12-31T23:59:59Z"). If no expiration date is supplied, the expiration date will default to null and
never expire.

## Example Usage

When a token has an expiry:

```hcl
resource "time_rotating" "example" {
  rotation_days = 30
}

resource "tfe_audit_trail_token" "test" {
  organization = data.tfe_organization.org.name
  expired_at = time_rotating.example.rotation_rfc3339
}
```

## Attributes Reference

* `id` - The ID of the token.
* `token` - The generated token.

## Import

Audit trail tokens can be imported; use `<ORGANIZATION NAME>` as the import ID.
For example:

```shell
terraform import tfe_audit_trail_token.test my-org-name
```
