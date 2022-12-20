---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_member"
description: |-
  Add or remove a user from a team.
---

# tfe_team_member

Add or remove a user from a team.

~> **NOTE** on managing team memberships: Terraform currently provides four
resources for managing team memberships.
The [tfe_team_organization_member](team_organization_member.html) and [tfe_team_organization_members](team_organization_members.html) resources are
the preferred way. The [tfe_team_member](team_member.html)
resource can be used multiple times as it manages the team membership for a
single user.  The [tfe_team_members](team_members.html) resource, on the other
hand, is used to manage all team memberships for a specific team and can only be
used once. All four resources cannot be used for the same team simultaneously.

## Example Usage

Basic usage:

```hcl
resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_team_member" "test" {
  team_id  = tfe_team.test.id
  username = "sander"
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `username` - (Required) Name of the user to add.

## Import

A team member can be imported; use `<TEAM ID>/<USERNAME>` as the import ID. For
example:

```shell
terraform import tfe_team_member.test team-47qC3LmA47piVan7/sander
```
