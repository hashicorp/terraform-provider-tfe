---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_membership"
sidebar_current: "docs-resource-tfe-organization-membership"
description: |-
  Add or remove a user from an organization.
---

# tfe_organization_membership

Add or remove a user from an organization.

~> **NOTE:** This resource requires using the provider with Terraform Cloud or
an instance of Terraform Enterprise at least as recent as v202004-1.

~> **NOTE:** This resource cannot be used to update an existing user's email address
since users themselves are the only ones permitted to update their email address.
If a user updates their email address, configurations using the email address should
be updated manually.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization_membership" "test" {
  organization  = "my-org-name"
  email = "user@company.com"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) Name of the organization.
* `email` - (Required) Email of the user to add.
