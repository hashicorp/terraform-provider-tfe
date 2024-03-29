---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_policy_set_parameter"
description: |-
  Manages policy set parameters.
---

# tfe_policy_set_parameter

Creates, updates and destroys policy set parameters.

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
            Name = "my-policy-set-name",
            Organization = Token.AsString(tfeOrganizationTest.Id)
        });
        /*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
        tfePolicySetTest.OverrideLogicalId("test");
        var tfePolicySetParameterTest =
        new PolicySetParameter.PolicySetParameter(this, "test_2", new PolicySetParameterConfig {
            Key = "my_key_name",
            PolicySetId = Token.AsString(tfePolicySetTest.Id),
            Value = "my_value_name"
        });
        /*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
        tfePolicySetParameterTest.OverrideLogicalId("test");
    }
}
```

## Argument Reference

The following arguments are supported:

* `Key` - (Required) Name of the parameter.
* `Value` - (Required) Value of the parameter.
* `Sensitive` - (Optional) Whether the value is sensitive. If true then the
  parameter is written once and not visible thereafter. Defaults to `False`.
* `PolicySetId` - (Required) The ID of the policy set that owns the parameter.

## Attributes Reference

* `Id` - The ID of the parameter.

## Import

Parameters can be imported; use
`<POLICY SET ID>/<PARAMETER ID>` as the import ID. For
example:

```shell
terraform import tfe_policy_set_parameter.test polset-wAs3zYmWAhYK7peR/var-5rTwnSaRPogw6apb
```


<!-- cache-key: cdktf-0.17.0-pre.15 input-3d439f538435c91fac393d64ba8c1ac4db8481770f20e794bdb2cde671211a74 -->