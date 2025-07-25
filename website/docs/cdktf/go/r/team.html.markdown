---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team"
description: |-
  Manages teams.
---


<!-- Please do not edit this file, it is generated. -->
# tfe_team

Manages teams.

## Example Usage

Basic usage:

```go
import constructs "github.com/aws/constructs-go/constructs"
import cdktf "github.com/hashicorp/terraform-cdk-go/cdktf"
/*Provider bindings are generated by running cdktf get.
See https://cdk.tf/provider-generation for more details.*/
import "github.com/aws-samples/dummy/gen/providers/tfe/team"
type myConvertedCode struct {
	terraformStack
}

func newMyConvertedCode(scope construct, name *string) *myConvertedCode {
	this := &myConvertedCode{}
	cdktf.NewTerraformStack_Override(this, scope, name)
	team.NewTeam(this, jsii.String("test"), &teamConfig{
		name: jsii.String("my-team-name"),
		organization: jsii.String("my-org-name"),
	})
	return this
}
```

Organization Permission usage:

```hcl
resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
  organization_access {
    manage_vcs_settings = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `Name` - (Required) Name of the team.
* `Organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `Visibility` - (Optional) The visibility of the team ("secret" or "organization")
* `OrganizationAccess` - (Optional) Settings for the team's [organization access](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/permissions#organization-permissions).
* `SsoTeamId` - (Optional) Unique Identifier to control [team membership](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/single-sign-on#team-names-and-sso-team-ids) via SAML. Defaults to `Null`
* `AllowMemberTokenManagement` - (Optional) Used by Owners and users with "Manage Teams" permissions to control whether team members can manage team tokens. Defaults to `True`.

The `OrganizationAccess` block supports:

* `ReadWorkspaces` - (Optional) Allow members to view all workspaces in this organization.
* `ReadProjects` - (Optional) Allow members to view all projects within the organization. Requires `ReadWorkspaces` to be set to `True`.
* `ManagePolicies` - (Optional) Allows members to create, edit, and delete the organization's Sentinel policies.
* `ManagePolicyOverrides` - (Optional) Allows members to override soft-mandatory policy checks.
* `ManageWorkspaces` - (Optional) Allows members to create and administrate all workspaces within the organization.
* `ManageVcsSettings` - (Optional) Allows members to manage the organization's VCS Providers and SSH keys.
* `ManageProviders` - (Optional) Allow members to publish and delete providers in the organization's private registry.
* `ManageModules` - (Optional) Allow members to publish and delete modules in the organization's private registry.
* `ManageRunTasks` - (Optional) Allow members to create, edit, and delete the organization's run tasks.
* `ManageProjects` - (Optional) Allow members to create and administrate all projects within the organization. Requires `ManageWorkspaces` to be set to `True`.
* `ManageMembership` - (Optional) Allow members to add/remove users from the organization, and to add/remove users from visible teams.
* `ManageTeams` - (Optional) Allow members to create, update, and delete teams.
* `ManageOrganizationAccess` - (Optional) Allow members to update the organization access settings of teams.
* `AccessSecretTeams` - (Optional) Allow members access to secret teams up to the level of permissions granted by their team permissions setting.
* `ManageAgentPools` - (Optional) Allow members to create, edit, and delete agent pools within their organization.

## Attributes Reference

* `Id` The ID of the team.

## Import

Teams can be imported; use `<ORGANIZATION NAME>/<TEAM ID>` or `<ORGANIZATION NAME>/<TEAM NAME>` as the import ID. For
example:

```shell
terraform import tfe_team.test my-org-name/team-uomQZysH9ou42ZYY
```
or
```shell
terraform import tfe_team.test my-org-name/my-team-name
```

<!-- cache-key: cdktf-0.17.0-pre.15 input-d9bc393198d4cc776e7fe4082a32d50cc258f18eded6123cefeca267b75b61a6 -->