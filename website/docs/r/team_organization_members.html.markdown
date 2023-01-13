---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_organization_members"
description: |-
  Add or remove users from a team based on their organization memberships.
---

# tfe_team_organization_members

Add or remove one or more team members using a
[tfe_organization_membership](organization_membership.html).

~> **NOTE** on managing team memberships: Terraform currently provides four
resources for managing team memberships. This - along with [tfe_team_organization_member](team_organization_member.html) - is the preferred method as it
allows you to add members to a team by email addresses. The [tfe_team_organization_member](team_organization_member.html) is used to manage a single team membership whereas [tfe_team_organization_members](team_organization_members.html) is used to manage all team memberships at once. All four resources cannot be used for the same team simultaneously.

~> **NOTE:** This resource requires using the provider with Terraform Cloud or
an instance of Terraform Enterprise at least as recent as v202004-1.

## Example Usage

Basic usage:

```hcl
resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_organization_membership" "test" {
  organization = "my-org-name"
  email = "example@hashicorp.com"
}

resource "tfe_organization_membership" "sample" {
  organization = "my-org-name"
  email = "sample@hashicorp.com"
}

resource "tfe_team_organization_members" "test" {
  team_id = tfe_team.test.id
  organization_membership_ids = [
    tfe_organization_membership.test.id,
    tfe_organization_membership.sample.id
  ]
}
```

With a set of organization members:

```hcl
locals {
  all_users = toset([
    "user1@hashicorp.com",
    "user2@hashicorp.com",
  ])
}

resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_organization_membership" "all_membership" {
  organization = "my-org-name"
  for_each     = local.all_users
  email        = each.key
}

resource "tfe_team_organization_members" "test" {
  team_id = tfe_team.test.id
  organization_membership_ids = [for member in local.all_users : tfe_organization_membership.all_membership[member].id]
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `organization_membership_ids` - (Required) IDs of organization memberships to be added.

## Import

A resource can be imported by using the team ID `<TEAM ID>`
as the import ID. For example:

```shell
terraform import tfe_team_organization_members.test team-47qC3LmA47piVan7
```
