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

```go
import constructs "github.com/aws/constructs-go/constructs"
import "github.com/hashicorp/terraform-cdk-go/cdktf"
/*Provider bindings are generated by running cdktf get.
See https://cdk.tf/provider-generation for more details.*/
import "github.com/aws-samples/dummy/gen/providers/tfe/team"
import "github.com/aws-samples/dummy/gen/providers/tfe/workspace"
import "github.com/aws-samples/dummy/gen/providers/tfe/teamAccess"
type myConvertedCode struct {
	terraformStack
}

func newMyConvertedCode(scope construct, name *string) *myConvertedCode {
	this := &myConvertedCode{}
	cdktf.NewTerraformStack_Override(this, scope, name)
	tfeTeamTest := team.NewTeam(this, jsii.String("test"), &teamConfig{
		name: jsii.String("my-team-name"),
		organization: jsii.String("my-org-name"),
	})
	tfeWorkspaceTest := workspace.NewWorkspace(this, jsii.String("test_1"), &workspaceConfig{
		name: jsii.String("my-workspace-name"),
		organization: jsii.String("my-org-name"),
	})
	/*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
	tfeWorkspaceTest.OverrideLogicalId(jsii.String("test"))
	tfeTeamAccessTest := teamAccess.NewTeamAccess(this, jsii.String("test_2"), &teamAccessConfig{
		access: jsii.String("read"),
		teamId: cdktf.Token_AsString(tfeTeamTest.id),
		workspaceId: cdktf.Token_*AsString(tfeWorkspaceTest.id),
	})
	/*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
	tfeTeamAccessTest.OverrideLogicalId(jsii.String("test"))
	return this
}
```

## Argument Reference

The following arguments are supported:

* `TeamId` - (Required) ID of the team to add to the workspace.
* `WorkspaceId` - (Required) ID of the workspace to which the team will be added.
* `Access` - (Optional) Type of fixed access to grant. Valid values are `Admin`, `Read`, `Plan`, or `Write`. To use `Custom` permissions, use a `Permissions` block instead. This value _must not_ be provided if `Permissions` is provided.
* `Permissions` - (Optional) Permissions to grant using [custom workspace permissions](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/permissions#custom-workspace-permissions). This value _must not_ be provided if `Access` is provided.

The `Permissions` block supports:

* `Runs` - (Required) The permission to grant the team on the workspace's runs. Valid values are `Read`, `Plan`, or `Apply`.
* `Variables` - (Required) The permission to grant the team on the workspace's variables. Valid values are `None`, `Read`, or `Write`.
* `StateVersions` - (Required) The permission to grant the team on the workspace's state versions. Valid values are `None`, `Read`, `ReadOutputs`, or `Write`.
* `SentinelMocks` - (Required) The permission to grant the team on the workspace's generated Sentinel mocks, Valid values are `None` or `Read`.
* `WorkspaceLocking` - (Required) Boolean determining whether or not to grant the team permission to manually lock/unlock the workspace.
* `RunTasks` - (Required) Boolean determining whether or not to grant the team permission to manage workspace run tasks.

-> **Note:** At least one of `Access` or `Permissions` _must_ be provided, but not both. Whichever is omitted will automatically reflect the state of the other.

## Attributes Reference

* `Id` The team access ID.

## Import

Team accesses can be imported; use
`<ORGANIZATION NAME>/<WORKSPACE NAME>/<TEAM ACCESS ID>` as the import ID. For
example:

```shell
terraform import tfe_team_access.test my-org-name/my-workspace-name/tws-8S5wnRbRpogw6apb
```

<!-- cache-key: cdktf-0.17.0-pre.15 input-1f416915697c1b047d62f590ef6bc829e7f2a7f58be51029af4020952110b5d6 -->