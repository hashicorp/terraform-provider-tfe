---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_policy_set_exclusion"
description: |-
  Add a policy set to an excluded workspace
---


<!-- Please do not edit this file, it is generated. -->
# tfe_workspace_policy_set_exclusion

Adds and removes policy sets from an excluded workspace

-> **Note:** `TfePolicySet` has an argument `WorkspaceIds` that should not be used alongside this resource. They attempt to manage the same attachments.

## Example Usage

Basic usage:

```csharp
using Constructs;
using HashiCorp.Cdktf;
/*Provider bindings are generated by running cdktf get.
See https://cdk.tf/provider-generation for more details.*/
using Gen.Providers.Tfe;
class MyConvertedCode : TerraformStack
{
    public MyConvertedCode(Construct scope, string name) : base(scope, name)
    {
        var tfeOrganizationTest = new Organization.Organization(this, "test", new OrganizationConfig {
            Email = "admin@company.com",
            Name = "my-org-name"
        });
        var tfePolicySetTest = new PolicySet.PolicySet(this, "test_1", new PolicySetConfig {
            Description = "Some description.",
            Name = "my-policy-set",
            Organization = Token.AsString(tfeOrganizationTest.Name)
        });
        /*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
        tfePolicySetTest.OverrideLogicalId("test");
        var tfeWorkspaceTest = new Workspace.Workspace(this, "test_2", new WorkspaceConfig {
            Name = "my-workspace-name",
            Organization = Token.AsString(tfeOrganizationTest.Name)
        });
        /*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
        tfeWorkspaceTest.OverrideLogicalId("test");
        var tfeWorkspacePolicySetExclusionTest =
        new WorkspacePolicySetExclusion.WorkspacePolicySetExclusion(this, "test_3", new WorkspacePolicySetExclusionConfig {
            PolicySetId = Token.AsString(tfePolicySetTest.Id),
            WorkspaceId = Token.AsString(tfeWorkspaceTest.Id)
        });
        /*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
        tfeWorkspacePolicySetExclusionTest.OverrideLogicalId("test");
    }
}
```

## Argument Reference

The following arguments are supported:

* `PolicySetId` - (Required) ID of the policy set.
* `WorkspaceId` - (Required) Excluded workspace ID to add the policy set to.

## Attributes Reference

* `Id` - The ID of the policy set attachment. ID format: `<workspace-id>_<policy-set-id>`

## Import

Excluded Workspace Policy Sets can be imported; use `<ORGANIZATION>/<WORKSPACE NAME>/<POLICY SET NAME>`. For example:

```shell
terraform import tfe_workspace_policy_set_exclusion.test 'my-org-name/workspace/policy-set-name'
```

<!-- cache-key: cdktf-0.17.0-pre.15 input-997093454ffecd4c222a0dd8330d635ee80e18f9571307774bda8fd64ae29fa2 -->