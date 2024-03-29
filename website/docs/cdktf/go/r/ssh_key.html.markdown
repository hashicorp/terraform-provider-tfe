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

```go
import constructs "github.com/aws/constructs-go/constructs"
import cdktf "github.com/hashicorp/terraform-cdk-go/cdktf"
/*Provider bindings are generated by running cdktf get.
See https://cdk.tf/provider-generation for more details.*/
import "github.com/aws-samples/dummy/gen/providers/tfe/sshKey"
type myConvertedCode struct {
	terraformStack
}

func newMyConvertedCode(scope construct, name *string) *myConvertedCode {
	this := &myConvertedCode{}
	cdktf.NewTerraformStack_Override(this, scope, name)
	sshKey.NewSshKey(this, jsii.String("test"), &sshKeyConfig{
		key: jsii.String("private-ssh-key"),
		name: jsii.String("my-ssh-key-name"),
		organization: jsii.String("my-org-name"),
	})
	return this
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