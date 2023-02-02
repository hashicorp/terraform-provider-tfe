---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_access"
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
  team_id      = tfe_team.test.id
  workspace_id = tfe_workspace.test.id
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team to add to the workspace.
* `workspace_id` - (Required) ID of the workspace to which the team will be added.
* `access` - (Optional) Type of fixed access to grant. Valid values are `admin`, `read`, `plan`, or `write`. To use `custom` permissions, use a `permissions` block instead. This value _must not_ be provided if `permissions` is provided.
* `permissions` - (Optional) Permissions to grant using [custom workspace permissions](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/permissions#custom-workspace-permissions). This value _must not_ be provided if `access` is provided.

The `permissions` block supports:

* `runs` - (Required) The permission to grant the team on the workspace's runs. Valid values are `read`, `plan`, or `apply`.
* `variables` - (Required) The permission to grant the team on the workspace's variables. Valid values are `none`, `read`, or `write`.
* `state_versions` - (Required) The permission to grant the team on the workspace's state versions. Valid values are `none`, `read`, `read-outputs`, or `write`.
* `sentinel_mocks` - (Required) The permission to grant the team on the workspace's generated Sentinel mocks, Valid values are `none` or `read`.
* `workspace_locking` - (Required) Boolean determining whether or not to grant the team permission to manually lock/unlock the workspace.
* `run_tasks` - (Required) Boolean determining whether or not to grant the team permission to manage workspace run tasks.

-> **Note:** At least one of `access` or `permissions` _must_ be provided, but not both. Whichever is omitted will automatically reflect the state of the other.

## Attributes Reference

* `id` The team access ID.

## Import

Team accesses can be imported; use
`<ORGANIZATION NAME>/<WORKSPACE NAME>/<TEAM ACCESS ID>` as the import ID. For
example:

```shell
terraform import tfe_team_access.test my-org-name/my-workspace-name/tws-8S5wnRbRpogw6apb
```
