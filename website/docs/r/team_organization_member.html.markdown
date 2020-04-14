---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_organization_member"
sidebar_current: "docs-resource-tfe-team-organization_member"
description: |-
  Add or remove a user from a team.
---

# tfe_team_organization_member

Add or remove a team member using a
[tfe_organization_membership](organization_membership.html).

~> **NOTE** on managing team memberships: Terraform currently provides three
resources for managing team memberships. This is the preferred method as it
allows you to add a member to a team by email address.

~> **NOTE:** This resource requires using the provider with Terraform Cloud or
an instance of Terraform Enterprise at least as recent as v202004-1.

## Example Usage

Basic usage:

```hcl
resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_organization_membership "test" {
  organization = "my-org-name"
  email = "example@hashicorp.com"
}

resource "tfe_team_organization_member" "test" {
  team_id = "${tfe_team.test.id}"
  organization_membership_id = "${tfe_organization_membership.test.id}"
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `organization_membership_id` - (Required) ID of the organization membership.

## Import

A team member can be imported; use `<TEAM ID>/<ORGANIZATION MEMBERSHIP ID>`
as the import ID. For example:

```shell
terraform import tfe_team_organization_member.test team-47qC3LmA47piVan7/ou-2342390sdf0jj
```
