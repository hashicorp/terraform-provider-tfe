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

```java
import software.constructs.*;
import com.hashicorp.cdktf.*;
/*Provider bindings are generated by running cdktf get.
See https://cdk.tf/provider-generation for more details.*/
import gen.providers.tfe.organization.*;
import gen.providers.tfe.policySet.*;
import gen.providers.tfe.policySetParameter.*;
public class MyConvertedCode extends TerraformStack {
    public MyConvertedCode(Construct scope, String name) {
        super(scope, name);
        Organization tfeOrganizationTest = new Organization(this, "test", new OrganizationConfig()
                .email("admin@company.com")
                .name("my-org-name")
                );
        PolicySet tfePolicySetTest = new PolicySet(this, "test_1", new PolicySetConfig()
                .name("my-policy-set-name")
                .organization(Token.asString(tfeOrganizationTest.getId()))
                );
        /*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
        tfePolicySetTest.overrideLogicalId("test");
        PolicySetParameter tfePolicySetParameterTest =
        new PolicySetParameter(this, "test_2", new PolicySetParameterConfig()
                .key("my_key_name")
                .policySetId(Token.asString(tfePolicySetTest.getId()))
                .value("my_value_name")
                );
        /*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
        tfePolicySetParameterTest.overrideLogicalId("test");
    }
}
```

## Argument Reference

The following arguments are supported:

* `key` - (Required) Name of the parameter.
* `value` - (Required) Value of the parameter.
* `sensitive` - (Optional) Whether the value is sensitive. If true then the
  parameter is written once and not visible thereafter. Defaults to `false`.
* `policySetId` - (Required) The ID of the policy set that owns the parameter.

## Attributes Reference

* `id` - The ID of the parameter.

## Import

Parameters can be imported; use
`<POLICY SET ID>/<PARAMETER ID>` as the import ID. For
example:

```shell
terraform import tfe_policy_set_parameter.test polset-wAs3zYmWAhYK7peR/var-5rTwnSaRPogw6apb
```


<!-- cache-key: cdktf-0.17.0-pre.15 input-3d439f538435c91fac393d64ba8c1ac4db8481770f20e794bdb2cde671211a74 -->