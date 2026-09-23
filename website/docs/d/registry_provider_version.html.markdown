---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_provider_version"
description: |-
  Get information on a private registry provider version.
---

# Data Source: tfe_registry_provider_version

Use this data source to get information about a private registry provider version.

## Example Usage

```hcl
data "tfe_registry_provider_version" "example" {
  organization = "my-org-name"
  name         = "my-provider"
  version      = "1.0.0"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `registry_name` - (Optional) Whether this is a publicly maintained provider or private. Must be either `public` or `private`. Defaults to `private`.
* `namespace` - (Optional) The namespace of the provider. For private providers this is the same as the organization.
* `name` - (Required) Name of the provider.
* `version` - (Required) The version of the provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the provider version.
* `key_id` - The GPG key ID used to sign the provider version.
* `protocols` - An array of Terraform provider API versions that this version supports.
* `shasums_uploaded` - Indicates whether the SHASUMS file has been uploaded.
* `shasums_sig_uploaded` - Indicates whether the SHASUMS signature file has been uploaded.
* `platforms` - List of platform release files for this provider version. Each platform has the following attributes:
  * `id` - ID of the provider version platform.
  * `os_arch` - Operating system and architecture (e.g., `linux_amd64`).
  * `filename` - The filename of the provider binary.
  * `shasum` - The SHA256 checksum of the provider binary.
  * `provider_binary_uploaded` - Indicates whether the provider binary has been uploaded.
* `created_at` - The time when the provider version was created.
* `updated_at` - The time when the provider version was last updated.