---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization"
sidebar_current: "docs-resource-tfe-organization-x"
description: |-
  Manages organizations.
---

# tfe_organization

Manages organizations.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name = "my-org-name"
  email = "admin@company.com"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the organization.
* `email` - (Required) Admin email address.
* `session_timeout_minutes` - (Optional) Session timeout after inactivity.
  Defaults to `20160`.
* `session_remember_minutes` - (Optional) Session expiration. Defaults to
  `20160`.
* `collaborator_auth_policy` - (Optional) Authentication policy (`password`
  or `two_factor_mandatory`). Defaults to `password`.

## Attributes Reference

* `id` - The name of the organization.

## Import

Organizations can be imported; use `<ORGANIZATION NAME>` as the import ID. For
example:

```shell
terraform import tfe_organization.test my-org-name
```
