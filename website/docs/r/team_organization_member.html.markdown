---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_organization_member"
description: |-
  Add or remove a user from a team.
---

# tfe_team_organization_member

Add or remove a team member using a
[tfe_organization_membership](organization_membership.html).

~> **NOTE** on managing team memberships: Terraform currently provides four
resources for managing team memberships. This - along with [tfe_organization_membership](organization_membership.html) - is the preferred method as it
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

resource "tfe_team_organization_member" "test" {
  team_id = tfe_team.test.id
  organization_membership_id = tfe_organization_membership.test.id
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `organization_membership_id` - (Required) ID of the organization membership.

## Import

A team member can be imported; use `<TEAM ID>/<ORGANIZATION MEMBERSHIP ID>` or `<ORGANIZATION NAME>/<USER EMAIL>/<TEAM NAME>`
as the import ID. For example:

```shell
terraform import tfe_team_organization_member.test team-47qC3LmA47piVan7/ou-2342390sdf0jj
```
or
```shell
terraform import tfe_team_organization_member.test my-org-name/user@company.com/my-team-name
```
~> **NOTE:** The `<ORGANIZATION NAME>/<USER EMAIL>/<TEAM NAME>` import ID format cannot be used if there are `/` characters in the user's email. These users must be imported with the `<TEAM ID>/<ORGANIZATION MEMBERSHIP ID>` format instead  