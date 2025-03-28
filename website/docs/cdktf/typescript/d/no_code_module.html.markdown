---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_no_code_module"
description: |-
  Get information on a no-code module.
---


<!-- Please do not edit this file, it is generated. -->
# Data Source: tfe_registry_provider

Use this data source to read the details of an existing No-Code-Allowed module.

## Example Usage

```typescript
// DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
import { Construct } from "constructs";
import { Token, TerraformStack } from "cdktf";
/*
 * Provider bindings are generated by running `cdktf get`.
 * See https://cdk.tf/provider-generation for more details.
 */
import { DataTfeNoCodeModule } from "./.gen/providers/tfe/data-tfe-no-code-module";
import { NoCodeModule } from "./.gen/providers/tfe/no-code-module";
class MyConvertedCode extends TerraformStack {
  constructor(scope: Construct, name: string) {
    super(scope, name);
    const foobar = new NoCodeModule(this, "foobar", {
      organization: Token.asString(tfeOrganizationFoobar.id),
      registryModule: Token.asString(tfeRegistryModuleFoobar.id),
    });
    const dataTfeNoCodeModuleFoobar = new DataTfeNoCodeModule(
      this,
      "foobar_1",
      {
        id: foobar.id,
      }
    );
    /*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
    dataTfeNoCodeModuleFoobar.overrideLogicalId("foobar");
  }
}

```

## Argument Reference

The following arguments are supported:

* `id` - (Required) ID of the no-code module. 

## Attributes Reference

* `id` - ID of the no-code module.
* `organization` - Organization name that the no-code module belongs to.
* `namespace` - Namespace name that the no-code module belongs to.
* `registryModuleId` - ID of the registry module for the no-code module. 
* `versionPin` - Version number the no-code module is pinned to.
* `enabled` - Indicates if this no-code module is currently enabled

<!-- cache-key: cdktf-0.20.8 input-575fd9c85b909c532a6abcfecdf6262d4ba6675f544b4a4a3a15b6c519eab693 -->