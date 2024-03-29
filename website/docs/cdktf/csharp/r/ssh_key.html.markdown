---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_ssh_key"
description: |-
  Manages SSH keys.
---

# tfe_ssh_key

This resource represents an SSH key which includes a name and the SSH private
key. An organization can have multiple SSH keys available.

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
        new SshKey.SshKey(this, "test", new SshKeyConfig {
            Key = "private-ssh-key",
            Name = "my-ssh-key-name",
            Organization = "my-org-name"
        });
    }
}
```

## Argument Reference

The following arguments are supported:

* `Name` - (Required) Name to identify the SSH key.
* `Organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `Key` - (Required) The text of the SSH private key.

## Attributes Reference

* `Id` The ID of the SSH key.

## Import

Because the Terraform Enterprise API does not return the private SSH key
content, this resource cannot be imported.

<!-- cache-key: cdktf-0.17.0-pre.15 input-ee94b4fd069224353c99784ca57ae132bbda89fc744065f36044f8e8c8a1f9b0 -->