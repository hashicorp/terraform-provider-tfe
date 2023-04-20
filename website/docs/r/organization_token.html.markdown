---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_token"
description: |-
  Generates a new organization token, replacing any existing token.
---

# tfe_organization_token

Generates a new organization token, replacing any existing token. This token
can be used to act as the organization service account.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization_token" "test" {
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `force_regenerate` - (Optional) If set to `true`, a new token will be
  generated even if a token already exists. This will invalidate the existing
  token!
* `expired_at` - (Optional) A date and time in which the organization token will expire. The expiration date must be passed in
  iso8601 format. If no expiration date is supplied, the expiration date will default to null and never expire.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization_token" "test" {
  organization = "my-org-name"
  expired_at = "2051-04-11T23:15:59+00:00"
}
```

## Attributes Reference

* `id` - The ID of the token.
* `token` - The generated token.

## Import

Organization tokens can be imported; use `<ORGANIZATION NAME>` as the import ID.
For example:

```shell
terraform import tfe_organization_token.test my-org-name
```
