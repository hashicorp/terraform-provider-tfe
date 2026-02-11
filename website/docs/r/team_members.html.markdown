---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_members"
description: |-
  Manages users in a team.
---

# tfe_team_members

Manages users in a team.

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

resource "tfe_team_members" "test" {
  team_id   = tfe_team.test.id
  usernames = ["admin", "sander"]
}
```

With a set of usernames:

```hcl
locals {
  all_usernames = toset([
    "user1",
    "user2",
  ])
}

resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_team_members" "test" {
  team_id   = tfe_team.test.id
  usernames = [for user in local.all_usernames : user]
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `usernames` - (Required) Names of the users to add.

## Attributes Reference

* `id` - The ID of the team.

## Import

Team members can be imported; use `<TEAM ID>` as the import ID. For example:

```shell
terraform import tfe_team_members.test team-47qC3LmA47piVan7
```
