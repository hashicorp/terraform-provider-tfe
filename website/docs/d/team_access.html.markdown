---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_access"
sidebar_current: "docs-datasource-tfe-team-access"
description: |-
  Get information on team permissions on a workspace.
---

# Data Source: tfe_team_access

Use this data source to get information about team permissions for a workspace.

## Example Usage

```hcl
data "tfe_team_access" "test" {
  team_id = "my-team-id"
  workspace_id = "my-workspace-id"
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `workspace_id` - (Required) ID of the workspace.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` The team access ID.
* `access` - The type of access granted.
