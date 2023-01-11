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

* `organization` - (Optional) Name of the organization. If omitted, default_organization provider config must be defined.
* `force_regenerate` - (Optional) If set to `true`, a new token will be
  generated even if a token already exists. This will invalidate the existing
  token!

## Attributes Reference

* `id` - The ID of the token.
* `token` - The generated token.

## Import

Organization tokens can be imported; use `<ORGANIZATION NAME>` as the import ID.
For example:

```shell
terraform import tfe_organization_token.test my-org-name
```
