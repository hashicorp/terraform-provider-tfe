---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_access"
description: |-
  Get information on team permissions on a workspace.
---


<!-- Please do not edit this file, it is generated. -->
# Data Source: tfe_team_access

Use this data source to get information about team permissions for a workspace.

## Example Usage

```python
# DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
from constructs import Construct
from cdktf import TerraformStack
#
# Provider bindings are generated by running `cdktf get`.
# See https://cdk.tf/provider-generation for more details.
#
from imports.tfe.data_tfe_team_access import DataTfeTeamAccess
class MyConvertedCode(TerraformStack):
    def __init__(self, scope, name):
        super().__init__(scope, name)
        DataTfeTeamAccess(self, "test",
            team_id="my-team-id",
            workspace_id="my-workspace-id"
        )
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `workspace_id` - (Required) ID of the workspace.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` The team access ID.
* `access` - The type of access granted to the team on the workspace.
* `permissions` - The permissions granted to the team on the workspaces for each whatever.

The `permissions` block contains:

* `runs` - The permission granted to runs. Valid values are `read`, `plan`, or `apply`
* `variables` - The permissions granted to variables. Valid values are `none`, `read`, or `write`
* `state_versions` - The permissions granted to state versions. Valid values are `none`, `read-outputs`, `read`, or `write`
* `sentinel_mocks` - The permissions granted to Sentinel mocks. Valid values are `none` or `read`
* `workspace_locking` - Whether permission is granted to manually lock the workspace or not.
* `run_tasks` - Boolean determining whether or not to grant the team permission to manage workspace run tasks.

<!-- cache-key: cdktf-0.20.8 input-95b20b8ad069cffffc1863d16cd9001f0074da34b788353d4b343912d7784d80 -->