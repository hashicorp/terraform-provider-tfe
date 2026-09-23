---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_provider_versions"
description: |-
  Get information on all versions of a private registry provider.
---

# Data Source: tfe_registry_provider_versions

Use this data source to get information about all versions of a private registry provider.

## Example Usage

```hcl
data "tfe_registry_provider_versions" "example" {
  organization = "my-org-name"
  name         = "my-provider"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `registry_name` - (Optional) Whether this is a publicly maintained provider or private. Must be either `public` or `private`. Defaults to `private`.
* `namespace` - (Optional) The namespace of the provider. For private providers this is the same as the organization.
* `name` - (Required) Name of the provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the data source (computed from organization, registry_name, namespace, and provider name).
* `versions` - List of provider versions. Each version has the following attributes:
  * `id` - ID of the provider version.
  * `organization` - Name of the organization.
  * `registry_name` - Whether this is a publicly maintained provider or private.
  * `namespace` - The namespace of the provider.
  * `name` - Name of the provider.
  * `version` - The version string.
  * `key_id` - The GPG key ID used to sign the provider version.
  * `protocols` - An array of Terraform provider API versions that this version supports.
  * `shasums_uploaded` - Indicates whether the SHASUMS file has been uploaded.
  * `shasums_sig_uploaded` - Indicates whether the SHASUMS signature file has been uploaded.
  * `created_at` - The time when the provider version was created.
  * `updated_at` - The time when the provider version was last updated.