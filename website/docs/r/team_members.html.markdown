---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_members"
sidebar_current: "docs-resource-tfe-team-members"
description: |-
  Manages users in a team.
---

# tfe_team_members

Manages users in a team. Users can be added by either their usernames or organization membership IDs.

~> **NOTE** on managing team memberships: Terraform currently provides three
resources for managing team memberships.
The [tfe_team_organization_member](team_organization_member.html) resource is
the preferred way. The [tfe_team_member](team_member.html)
resource can be used multiple times as it manages the team membership for a
single user.  The [tfe_team_members](team_members.html) resource, on the other
hand, is used to manage all team memberships for a specific team and can only be
used once. All three resources cannot be used for the same team simultaneously.

## Example Usage

Basic usage with usernames:

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

Basic usage with organization membership IDs:

```hcl
resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_organization_membership" "test" {
  email        = "test@mycompany.com"
  organization = "mycompany"
}

resource "tfe_team_members" "test" {
  team_id                     = tfe_team.test.id
  organization_membership_ids = [tfe_organization_membership.test.id]
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `usernames` - (Optional) Names of the users to add. Exactly one of `usernames` and `organization_membership_ids` has to be specified. They can not be used in conjuntion.
* `organization_membership_ids` - (Optional) Organization membership IDs of the users to add. Exactly one of `usernames` and `organization_membership_ids` has to be specified. They can not be used in conjuntion.

## Attributes Reference

* `id` - The ID of the team.

## Import

Team members can be imported; use `<TEAM ID>` as the import ID. For example:

```shell
terraform import tfe_team_members.test team-47qC3LmA47piVan7
```
