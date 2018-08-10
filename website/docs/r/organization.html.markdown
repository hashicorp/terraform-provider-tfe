---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization"
sidebar_current: "docs-tfe-organization"
description: |-
  Creates, updates and destroys organizations.
---

# tfe_organization

Creates, updates and destroys organizations.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "organization" {
	name = "my-org-name"
  email = "admin@company.com"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the organization.
* `email` - (Required) Admin email address.
* `session_timeout` - (Optional) Session timeout after inactivity (minutes).
  Defaults to 20160.
* `session_remember` - (Optional) Session expiration (minutes). Defaults to
  20160.
* `collaborator_auth_policy` - (Optional) Authentication policy (`password`
  or `two_factor_mandatory`). Defaults to `password`.

## Attributes Reference

* `id` - The name of the organization.
