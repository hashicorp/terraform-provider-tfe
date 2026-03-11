---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_current_user"
description: |-
  Get information on the current user associated with the API token.
---

# Data Source: tfe_current_user

Use this data source to get information about the current user associated with the API token used to configure the provider. When authenticated with a team or organization token, HCP Terraform returns a synthetic service user rather than a real user account, so attributes like `email` and `username` will not reflect a real person.

## Example Usage

```hcl
data "tfe_current_user" "current" {}

output "email" {
  value = data.tfe_current_user.current.email
}
```

A common use case is dynamically referencing the invoking user, for example when bootstrapping organization membership to avoid a conflict with the implicit owner:

```hcl
data "tfe_current_user" "current" {}

resource "tfe_organization_membership" "owner" {
  organization = "my-org"
  email        = data.tfe_current_user.current.email
}
```

## Argument Reference

This data source does not require any arguments.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `id` - The ID of the user.

* `username` - The username of the current user.

* `email` - The email address of the current user.

* `avatar_url` - Avatar URL of the current user.
