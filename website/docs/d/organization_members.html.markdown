---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_members"
sidebar_current: "docs-datasource-tfe-organization-members"
description: |-
  Get information on an Organization members.
---

# Data Source: tfe_organization_members

Use this data source to get information about members of an organization.

## Example Usage

```hcl
data "tfe_organization_members" "foo" {
  organization = "organization-name"
}
```

## Argument Reference

The following arguments are supported:
* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Name of the organization.
* `members` - A list of active members of the organization.
* `members_waiting` - A list of members with pending invite to organization.

The `member` block contains:

* `user_id` - The ID of the user.
* `membership_id` - The ID of the membership.