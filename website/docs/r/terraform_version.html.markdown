---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_terraform_version"
sidebar_current: "docs-resource-tfe-terraform-version-x"
description: |-
  Manages Terraform versions
---

# tfe_terraform_version

Manage Terraform versions available on Terraform Cloud/Enterprise.

## Example Usage

Basic Usage:

```hcl
resource "tfe_terraform_version" "test" {
  version = "1.1.2-custom"
  url = "https://tfe-host.com/path/to/terraform.zip"
  sha = "e75ac73deb69a6b3aa667cb0b8b731aee79e2904"
}
```

## Argument Reference

The following arguments are supported:

* `version` - (Required) A semantic version string in N.N.N or N.N.N-bundleName format.
* `url` - (Required) The URL where a ZIP-compressed 64-bit Linux binary of this version can be downloaded.
* `sha` - (Required) The SHA-256 checksum of the compressed Terraform binary.
* `official` - (Optional) Whether or not this is an official release of Terraform. Defaults to "false".
* `enabled` - (Optional) Whether or not this version of Terraform is enabled for use in Terraform Cloud/Enterprise. Defaults to "true".
* `beta` - (Optional) Whether or not this version of Terraform is beta pre-release. Defaults to "false".

## Attributes Reference

* `id` The ID of the Terraform version

## Import

Terraform versions can be imported; use `<TERRAFORM VERSION ID>` as the import ID. For example:

```shell
terraform import tfe_terraform_version.test tool-L4oe7rNwn7J4E5Yr 
```

-> **Note:** You can fetch a Terraform version ID from the URL of an exisiting version in the Terraform Cloud UI. The ID is in the format `tool-<RANDOM STRING>` 
