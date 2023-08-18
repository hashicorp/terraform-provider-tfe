---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_project_access"
description: |-
  Associate a team to permissions on a project.
---

# tfe_team_project_access

Associate a team to permissions on a project.

## Example Usage

Basic usage:

```hcl
resource "tfe_team" "admin" {
  name         = "my-admin-team"
  organization = "my-org-name"
}

resource "tfe_project" "test" {
  name         = "myproject"
  organization = "my-org-name"
}

resource "tfe_team_project_access" "admin" {
  access       = "admin"
  team_id      = tfe_team.admin.id
  project_id   = tfe_project.test.id
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team to add to the project.
* `project_id` - (Required) ID of the project to which the team will be added.
* `access` - (Required) Type of fixed access to grant. Valid values are `admin`, `maintain`, `write`, `read`, or `custom`.

## Custom Access

If using `custom` for `access`, you can set the levels of individual permissions
that affect the project itself and all workspaces in the project, by using `project_access` and `workspace_access` arguments and their associated permission attributes. When using custom access, if attributes are not set they will be given a default value. Some permissions have values that are specific "strings" that denote the level of the permission, while other permissions are simple booleans.

The following permissions apply to the project itself.

| project_access      | Description, Default, Valid Values          |
|---------------------|---------------------------------------------|
| `settings`          | The permission to grant for the project's settings. Default: `read`. Valid strings: `read`, `update`, or `delete` |
| `teams`             | The permission to grant for the project's teams. Default: `none`, Valid strings: `none`, `read`, or `manage` |

</n>
</n>
</n>

The following permissions apply to all workpsaces (and future workspaces) in the project.

| workspace_access     | Description, Default, Valid Values                    |
|----------------------|-------------------------------------------------------|
| `runs`               | The permission to grant project's workspaces' runs. Default: `read`. Valid strings: `read`, `plan`, or `apply`. |
| `sentinel_mocks`     | The permission to grant project's workspaces' Sentinel mocks. Default: `none`. Valid strings: `none`, or `read`. |
| `state_versions`     | The permission to grant project's workspaces' state versions. Default: `none` Valid strings: `none`, `read-outputs`, `read`, or `write`.|
| `variables`          | The permission to grant project's workspaces' variables. Default `none`. Valid strings: `none`, `read`, or `write`. |
| `create`             | The permission to create project's workspaces in the project. Default: `false`. Valid booleans `true`, `false` |
| `locking`            | The permission to manually lock or unlock the project's workspaces. Default `false`. Valid booleans `true`, `false` |
| `delete`             | The permission to delete the project's workspaces. Default: `false`. Valid booleans: `true`, `false` |
| `move`               | This permission to move workspaces into and out of the project. The team must also have permissions to the project(s) receiving the the workspace(s). Default: `false`. Valid booleans: `true`, `false` |
| `run_tasks`          | The permission to manage run tasks within the project's workspaces. Default `false`. Valid booleans: `true`, `false` |


## Example Usage with Custom Project Permissions

```hcl
resource "tfe_team" "dev" {
  name         = "my-dev-team"
  organization = "my-org-name"
}

resource "tfe_project" "test" {
  name         = "myproject"
  organization = "my-org-name"
}

resource "tfe_team_project_access" "custom" {
  access       = "custom"
  team_id      = tfe_team.dev.id
  project_id   = tfe_project.test.id

  project_access {
    settings = "read"
    teams    = "none"
  }
  workspace_access {
    state_versions = "write"
    sentinel_mocks = "none"
    runs           = "apply"
    variables      = "write"
    create         = true
    locking        = true
    move           = false
    delete         = false
    run_tasks      = false
  }
}
```

## Attributes Reference

* `id` The team project access ID.

## Import

Team project accesses can be imported; use the project team access ID as the import ID. For
example:

```shell
terraform import tfe_team_project_access.admin tprj-2pmtXpZa4YzVMTPi
```
