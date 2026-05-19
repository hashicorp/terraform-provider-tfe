---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_provider_version"
description: |-
  Manages private registry provider versions.
---

# tfe_registry_provider_version

Manages a version of a Provider Version in a private registry. Unlike the public Terraform Registry, the private registry does not automatically upload new releases. You must manually add new provider versions and the associated release files.

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
  
  shasums_filename     = "./path/to/terraform-provider-example-1.0.0-SHA256SUMS"
  shasums_sig_filename = "./path/to/terraform-provider-example-1.0.0-SHA256SUMS.sig"
}
```

## Argument Reference

The following arguments are supported:

* `version` - (Required) A valid semver version string.
* `key_id` - (Required) The Key ID from a GPG key managed by `tfe_registry_gpg_key` or otherwise uploaded to the organization.
* `protocols` - (Required) An array of Terraform provider API versions that this version supports. Must be one or all of the following values ["4.0","5.0","6.0"].
* `shasums_filename` - (Required) The path to the SHA256SUMS file describing the version release signing.
* `shasums_sig_filename` - (Required) The path to the SHA256SUMS.sig file describing the version release signing.

## Attributes Reference

* `id` - ID of the provider version.
* `created_at` - The time when the provider version was created.
* `updated_at` - The time when the provider version was last updated.
* `permissions` - A block with read-only permissions related to this provider version

The `permissions` block supports the following attributes:

* `can-delete` - Whether or not the provider version can be destroyed

## Import

Keys can be imported; use `<registry provider id>` as the import ID. For
example:

```shell
terraform import tfe_registry_provider_version.example provver-y5KZUsSBRLV9zCtL
```
