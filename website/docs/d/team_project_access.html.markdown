---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_project_access"
description: |-
  Get information on team permissions on a project.
---

# Data Source: tfe_team_project_access

Use this data source to get information about team permissions for a project.

## Example Usage

```hcl
data "tfe_team_project_access" "test" {
  team_id    = "my-team-id"
  project_id = "my-project-id"
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `project_id` - (Required) ID of the project.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The team project access ID.
* `access` - The type of access granted to the team on the project.
* `project_access` - The permissions granted to the team on the project itself.
* `workspace_access` - The permissions granted to the team across all workspaces in the project.

The `project_access` block contains:

* `settings` - The permission granted to the project's settings. Valid values are `read`, `update`, or `delete`.
* `teams` - The permission granted to the project's teams. Valid values are `none`, `read`, or `manage`.
* `variable_sets` - The permission granted to the project's variable sets. Valid values are `none`, `read`, or `write`.

The `workspace_access` block contains:

* `runs` - The permission granted to runs. Valid values are `read`, `plan`, or `apply`.
* `variables` - The permission granted to variables. Valid values are `none`, `read`, or `write`.
* `state_versions` - The permission granted to state versions. Valid values are `none`, `read-outputs`, `read`, or `write`.
* `sentinel_mocks` - The permission granted to Sentinel mocks. Valid values are `none` or `read`.
* `create` - Whether the team can create workspaces in the project.
* `locking` - Whether the team can manually lock or unlock workspaces in the project.
* `move` - Whether the team can move workspaces into and out of the project.
* `delete` - Whether the team can delete workspaces in the project.
* `run_tasks` - Whether the team can manage run tasks in the project's workspaces.
* `policy_overrides` - This permission allows a team to override soft-mandatory policy evaluations, provided that team has been granted the org level 'delegate policy overrides' permission.
