---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_access"
description: |-
  Get information on team permissions on a workspace.
---

# Data Source: tfe_team_access

Use this data source to get information about team permissions for a workspace.

## Example Usage

```hcl
data "tfe_team_access" "test" {
  team_id      = "my-team-id"
  workspace_id = "my-workspace-id"
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `workspace_id` - (Required) ID of the workspace.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The team access ID.
* `access` - The type of access granted to the team on the workspace.
* `permissions` - The custom permissions granted to the team on the workspace.

The `permissions` block contains:

* `runs` - The permission granted to runs. Valid values are `read`, `plan`, or `apply`.
* `variables` - The permission granted to variables. Valid values are `none`, `read`, or `write`.
* `state_versions` - The permission granted to state versions. Valid values are `none`, `read-outputs`, `read`, or `write`.
* `sentinel_mocks` - The permission granted to Sentinel mocks. Valid values are `none` or `read`.
* `workspace_locking` - Whether the team can manually lock or unlock the workspace.
* `run_tasks` - Whether the team can manage workspace run tasks.
* `policy_overrides` - This permission allows a team to override soft-mandatory policy evaluations, provided that team has been granted the org level 'delegate policy overrides' permission.
