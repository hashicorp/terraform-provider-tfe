---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_members"
sidebar_current: "docs-tfe-team-members"
description: |-
  Add or remove a users from a team.
---

# tfe_team_members

Add or remove a users from a team.

~> NOTE on managing team memberships: Terraform currently provides two resources
for managing team memberships. The [tfe_team_member](team_member.html) resource
can be used multiple times as it manages the team membership for a single user.
The [tfe_team_members](team_members.html) resource, on the other hand, is used
to manage all team memberships for a specific team and can only be used once.
Both resources cannot be used for the same team simultaneously.

## Example Usage

Basic usage:

```hcl
resource "tfe_team" "team" {
  name = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_team_members" "members" {
	team_id = "${tfe_team.team.id}"
  usernames = ["admin", "sander"]
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `usernames` - (Required) Names of the users to add.

## Attributes Reference

* `id` - The ID of the team.
