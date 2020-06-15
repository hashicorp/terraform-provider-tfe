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

* `team_id` - (Required) ID of the team to add to the workspace.
* `workspace_id` - (Required) ID of the workspace to which the team will be added.
* `access` - (Conflicts with `permissions`) Type of fixed access to grant. Valid values are `admin`, `read`, `plan`, or `write`. To use `custom` permissions, use a `permissions` block instead.
* `permissions` - (Conflicts with `access`) Permissions to grant using [custom workspace permissions](https://www.terraform.io/docs/cloud/users-teams-organizations/permissions.html#custom-workspace-permissions).

  The arguments for this block are:

  - `runs` - (Required) The permission to grant the team on the workspace's runs. Valid values are `read`, `plan`, or `apply`.
  - `variables` - (Required) The permission to grant the team on the workspace's variables. Valid values are `none`, `read`, or `write`.
  - `state_versions` - (Required) The permission to grant the team on the workspace's state versions. Valid values are `none`, `read`, `read-outputs`, or `write`.
  - `sentinel_mocks` - (Required) The permission to grant the team on the workspace's generated Sentinel mocks, Valid values are `none` or `read`.
  - `workspace_locking` - (Required) Boolean determining whether or not to grant the team permission to manually lock/unlock the workspace.

At least one of `access` or `permissions` must be specified, but not both. Whichever is omitted will automatically reflect the state of the other.

## Attributes Reference

* `id` The team access ID.

## Import

Team accesses can be imported; use
`<ORGANIZATION NAME>/<WORKSPACE NAME>/<TEAM ACCESS ID>` as the import ID. For
example:

```shell
terraform import tfe_team_access.test my-org-name/my-workspace-name/tws-8S5wnRbRpogw6apb
```
