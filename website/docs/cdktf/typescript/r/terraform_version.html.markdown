---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_terraform_version"
description: |-
  Manages Terraform versions
---


<!-- Please do not edit this file, it is generated. -->
# tfe_terraform_version

Manage Terraform versions available on HCP Terraform and Terraform Enterprise.

## Example Usage

Basic Usage:

```typescript
// DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
import { Construct } from "constructs";
import { TerraformStack } from "cdktf";
/*
 * Provider bindings are generated by running `cdktf get`.
 * See https://cdk.tf/provider-generation for more details.
 */
import { TerraformVersion } from "./.gen/providers/tfe/terraform-version";
class MyConvertedCode extends TerraformStack {
  constructor(scope: Construct, name: string) {
    super(scope, name);
    new TerraformVersion(this, "test", {
      sha: "e75ac73deb69a6b3aa667cb0b8b731aee79e2904",
      url: "https://tfe-host.com/path/to/terraform.zip",
      version: "1.1.2-custom",
    });
  }
}

```

## Argument Reference

The following arguments are supported:

* `version` - (Required) A semantic version string in N.N.N or N.N.N-bundleName format.
* `url` - (Required) The URL where a ZIP-compressed 64-bit Linux binary of this version can be downloaded.
* `sha` - (Required) The SHA-256 checksum of the compressed Terraform binary.
* `official` - (Optional) Whether or not this is an official release of Terraform. Defaults to "false".
* `enabled` - (Optional) Whether or not this version of Terraform is enabled for use in HCP Terraform and Terraform Enterprise. Defaults to "true".
* `beta` - (Optional) Whether or not this version of Terraform is beta pre-release. Defaults to "false".
* `deprecated` - (Optional) Whether or not this version of Terraform is deprecated. Defaults to "false".
* `deprecatedReason` - (Optional) Additional context about why a version of Terraform is deprecated. Defaults to "null" unless `deprecated` is true.

## Attributes Reference

* `id` The ID of the Terraform version

## Import

Terraform versions can be imported; use `<TERRAFORM VERSION ID>` or `<TERRAFORM VERSION NUMBER>` as the import ID. For example:

```shell
terraform import tfe_terraform_version.test tool-L4oe7rNwn7J4E5Yr
```

```shell
terraform import tfe_terraform_version.test 1.1.2
```

-> **Note:** You can fetch a Terraform version ID from the URL of an existing version in the HCP Terraform UI. The ID is in the format `tool-<RANDOM STRING>`

<!-- cache-key: cdktf-0.20.8 input-93d1722805fa12757f839282cbc6353a9d0aafa4f011bf5b15cbde86e59a32ef -->