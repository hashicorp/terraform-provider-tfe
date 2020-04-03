---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_membership"
sidebar_current: "docs-resource-tfe-organization-membership"
description: |-
  Add or remove a user from an organization.
---

# tfe_organization_membership

Add or remove a user from an organization.

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
