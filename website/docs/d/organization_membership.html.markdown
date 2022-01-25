---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_membership"
sidebar_current: "docs-datasource-tfe-organization-membership"
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

```hcl
data "tfe_organization_membership" "test" {
  organization  = "my-org-name"
  email = "user@company.com"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) Name of the organization.
* `email` - (Required) Email of the user.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The organization membership ID.
* `user_id` - The ID of the user associated with the organization membership.
* `username` - The Username of the user associated with the organization membership.
