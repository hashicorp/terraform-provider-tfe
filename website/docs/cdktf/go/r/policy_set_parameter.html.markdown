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

```go
import constructs "github.com/aws/constructs-go/constructs"
import "github.com/hashicorp/terraform-cdk-go/cdktf"
/*Provider bindings are generated by running cdktf get.
See https://cdk.tf/provider-generation for more details.*/
import "github.com/aws-samples/dummy/gen/providers/tfe/organization"
import "github.com/aws-samples/dummy/gen/providers/tfe/policySet"
import "github.com/aws-samples/dummy/gen/providers/tfe/policySetParameter"
type myConvertedCode struct {
	terraformStack
}

func newMyConvertedCode(scope construct, name *string) *myConvertedCode {
	this := &myConvertedCode{}
	cdktf.NewTerraformStack_Override(this, scope, name)
	tfeOrganizationTest := organization.NewOrganization(this, jsii.String("test"), &organizationConfig{
		email: jsii.String("admin@company.com"),
		name: jsii.String("my-org-name"),
	})
	tfePolicySetTest := policySet.NewPolicySet(this, jsii.String("test_1"), &policySetConfig{
		name: jsii.String("my-policy-set-name"),
		organization: cdktf.Token_AsString(tfeOrganizationTest.id),
	})
	/*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
	tfePolicySetTest.OverrideLogicalId(jsii.String("test"))
	tfePolicySetParameterTest :=
	policySetParameter.NewPolicySetParameter(this, jsii.String("test_2"), &policySetParameterConfig{
		key: jsii.String("my_key_name"),
		policySetId: cdktf.Token_*AsString(tfePolicySetTest.id),
		value: jsii.String("my_value_name"),
	})
	/*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
	tfePolicySetParameterTest.OverrideLogicalId(jsii.String("test"))
	return this
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