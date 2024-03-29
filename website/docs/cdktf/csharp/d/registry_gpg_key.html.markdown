---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_gpg_key"
description: |-
  Get information on a private registry GPG key.
---


<!-- Please do not edit this file, it is generated. -->
# Data Source: tfe_registry_gpg_key

Use this data source to get information about a private registry GPG key.

## Example Usage

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
        new DataTfeRegistryGpgKey.DataTfeRegistryGpgKey(this, "example", new DataTfeRegistryGpgKeyConfig {
            Id = "13DFECCA3B58CE4A",
            Organization = "my-org-name"
        });
    }
}
```

## Argument Reference

The following arguments are supported:

* `Id` - (Required) ID of the GPG key.
* `Organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.

## Attributes Reference

* `AsciiArmor` - ASCII-armored representation of the GPG key.
* `CreatedAt` - The time when the GPG key was created.
* `UpdatedAt` - The time when the GPG key was last updated.

<!-- cache-key: cdktf-0.17.0-pre.15 input-7cf721398cc48785bd0ab8f949360d917b2cadf37b1f704b8747ee2c07ced5d4 -->