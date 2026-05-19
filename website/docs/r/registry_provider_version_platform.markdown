---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_provider_version_platform"
description: |-
  Manages private registry provider versions.
---

# tfe_registry_provider_version_platform

Manages a platform release binary of a Provider Version in a private registry. Unlike the public Terraform Registry, the private registry does not automatically upload new releases. You must manually add new provider versions and the associated release files for each platform.

## Example Usage

```hcl
resource "tfe_registry_gpg_key" "my-key" {
  organization = "example"
  ascii_armor  = file("./path/to/my_key.pgp")
}

resource "tfe_registry_provider_version" "example" {
  version              = "3.1.1"
  key_id               = tfe_registry_gpg_key.my-key.id
  protocols            = ["5.0"]
  
  shasums_filename     = "./path/to/terraform-provider-example_1.0.0_SHA256SUMS"
  shasums_sig_filename = "./path/to/terraform-provider-example_1.0.0_SHA256SUMS.sig"
}

resource "tfe_registry_provider_version_platform" "example_linux" {
  os_arch  = "linux_amd64"
  filename = "./path/to/terraform-provider-example_1.0.0_linux_amd64.zip"
}

resource "tfe_registry_provider_version_platform" "example_darwin" {
  os_arch  = "darwin_amd64"
  filename = "./path/to/terraform-provider-example_1.0.0_darwin_amd64.zip"
}

resource "tfe_registry_provider_version_platform" "example_darwin_arm" {
  os_arch  = "darwin_arm64"
  filename = "./path/to/terraform-provider-example_1.0.0_darwin_arm64.zip"
}
```

## Argument Reference

The following arguments are supported:

* `os_arch` - (Required) A valid operating system string and architecture string, separated by an underscore. Example: `linux_amd64`.
* `filename` - (Required) The path to the release file described by the version signing defined in the associated `tfe_registry_provider_version`

## Attributes Reference

* `id` - ID of the provider version platform.
* `permissions` - A block with read-only permissions related to this provider version platform

The `permissions` block supports the following attributes:

* `can-delete` - Whether or not the provider version platform can be destroyed

## Import

Keys can be imported; use `<registry provider id>` as the import ID. For
example:

```shell
terraform import tfe_registry_provider_version.example provpltfrm-BLJWvWyJ2QMs525k
```
