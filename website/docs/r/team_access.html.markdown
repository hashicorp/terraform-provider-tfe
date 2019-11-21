---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_access"
sidebar_current: "docs-resource-tfe-team-access"
description: |-
  Associate a team to permissions on a workspace.
---

# tfe_team_access

Associate a team to permissions on a workspace.

## Example Usage

Basic usage:

```hcl
resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = "my-org-name"
}

resource "tfe_team_access" "test" {
  access       = "read"
  team_id      = "${tfe_team.test.id}"
  workspace_id = "${tfe_workspace.test.id}"
}
```

## Argument Reference

The following arguments are supported:

* `access` - (Required) Type of access to grant. Valid values are `admin`,
  `read`, `plan`, or `write`.
* `team_id` - (Required) ID of the team to add to the workspace.
* `workspace_id` - (Required) The workspace to which the team will be added,
  specified as the workspace's external id (`ws-123`).

## Attributes Reference

* `id` The team access ID.

## Import

Team accesses can be imported; use
`<ORGANIZATION NAME>/<WORKSPACE NAME>/<TEAM ACCESS ID>` as the import ID. For
example:

```shell
terraform import tfe_team_access.test my-org-name/my-workspace-name/tws-8S5wnRbRpogw6apb
```
