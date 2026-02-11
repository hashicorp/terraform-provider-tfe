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
  team_id      = "my-team-id"
  project_id   = "my-project-id"
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `project_id` - (Required) ID of the project.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` The team project access ID.
* `access` - The type of access granted to the team on the project.
