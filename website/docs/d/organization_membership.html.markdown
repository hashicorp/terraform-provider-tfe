---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_membership"
description: |-
  Get information on an organization membership.
---

# Data Source: tfe_organization_membership

Use this data source to get information about an organization membership.

~> **NOTE:** This data source requires using the provider with Terraform Cloud or
an instance of Terraform Enterprise at least as recent as v202004-1.

~> **NOTE:** If a user updates their email address, configurations using the email address should
be updated manually.

## Example Usage

### Fetch by email

```hcl
data "tfe_organization_membership" "test" {
  organization  = "my-org-name"
  email = "user@company.com"
}
```

### Fetch by username

```
data "tfe_organization_membership" "test" {
  organization  = "my-org-name"
  username = "my-username"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) Name of the organization.
* `email` - (Optional) Email of the user.
* `username` - (Optional) The username of the user.

~> **NOTE:** While `email` and `username` are optional arguments, one or the other is required.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The organization membership ID.
* `user_id` - The ID of the user associated with the organization membership.
* `username` - The username of the user associated with the organization membership.
