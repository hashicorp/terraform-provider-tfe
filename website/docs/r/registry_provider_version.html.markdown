---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_provider_version"
description: |-
  Manages private registry provider versions.
---

# tfe_registry_provider_version

Manages a version of a provider in the private registry. Unlike the public Terraform Registry, the private registry does not automatically upload new releases. You must manually add new provider versions and the associated release files.

## Example Usage

Create a private provider version:

```hcl
resource "tfe_organization" "example" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_provider" "example" {
  organization = tfe_organization.example.name
  name         = "my-provider"
}

resource "tfe_registry_gpg_key" "example" {
  organization = tfe_organization.example.name
  ascii_armor  = file("./path/to/my_key.pgp")
}

resource "tfe_registry_provider_version" "example" {
  organization = tfe_organization.example.name
  name         = tfe_registry_provider.example.name
  version      = "1.0.0"
  key_id       = tfe_registry_gpg_key.example.id
  protocols    = ["5.0"]
}
```

Create a private provider version with SHASUMS files:

```hcl
resource "tfe_organization" "example" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_provider" "example" {
  organization = tfe_organization.example.name
  name         = "my-provider"
}

resource "tfe_registry_gpg_key" "example" {
  organization = tfe_organization.example.name
  ascii_armor  = file("./path/to/my_key.pgp")
}

resource "tfe_registry_provider_version" "example" {
  organization     = tfe_organization.example.name
  name             = tfe_registry_provider.example.name
  version          = "1.0.0"
  key_id           = tfe_registry_gpg_key.example.id
  protocols        = ["5.0"]
  shasums_file     = "https://releases.example.com/my-provider/1.0.0/my-provider_1.0.0_SHA256SUMS"
  shasums_sig_file = "https://releases.example.com/my-provider/1.0.0/my-provider_1.0.0_SHA256SUMS.sig"
}
```

Create a public provider version:

```hcl
resource "tfe_organization" "example" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_provider" "example" {
  organization  = tfe_organization.example.name
  registry_name = "public"
  namespace     = "hashicorp"
  name          = "aws"
}

resource "tfe_registry_gpg_key" "example" {
  organization = tfe_organization.example.name
  ascii_armor  = file("./path/to/hashicorp_key.pgp")
}

resource "tfe_registry_provider_version" "example" {
  organization  = tfe_organization.example.name
  registry_name = "public"
  namespace     = "hashicorp"
  name          = "aws"
  version       = "5.0.0"
  key_id        = tfe_registry_gpg_key.example.id
  protocols     = ["5.0", "6.0"]
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `registry_name` - (Optional) Whether this is a publicly maintained provider or private. Must be either `public` or `private`. Defaults to `private`.
* `namespace` - (Optional) The namespace of the provider. Required if `registry_name` is `public`, otherwise it can't be configured, and it will be set to same value as the `organization`.
* `name` - (Required) Name of the provider.
* `version` - (Required) A valid semver version string.
* `key_id` - (Required) The GPG key ID used to sign the provider version. This must reference a GPG key that has been uploaded to the organization.
* `protocols` - (Required) An array of Terraform provider API versions that this version supports. Valid values are `"4.0"`, `"5.0"`, and `"6.0"`.
* `shasums_file` - (Optional) Path to the SHASUMS file (local file path or HTTP/HTTPS URL). If not provided, SHASUMS must be uploaded separately using the API.
* `shasums_sig_file` - (Optional) Path to the SHASUMS signature file (local file path or HTTP/HTTPS URL). If not provided, the signature must be uploaded separately using the API.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the provider version.
* `shasums_uploaded` - Indicates whether the SHASUMS file has been uploaded.
* `shasums_sig_uploaded` - Indicates whether the SHASUMS signature file has been uploaded.
* `created_at` - The time when the provider version was created.
* `updated_at` - The time when the provider version was last updated.

## Import

Provider versions can be imported using an identity. For example:

```hcl
import {
  to = tfe_registry_provider_version.example
  identity = {
    id            = "provver-y5KZUsSBRLV9zCtL"
    organization  = "my-org-name"
    registry_name = "private"
    namespace     = "my-org-name"
    name          = "my-provider"
    version       = "1.0.0"
    hostname      = "app.terraform.io"
  }
}
```

Provider versions can be imported using the Terraform CLI; use `<ORGANIZATION>/<REGISTRY NAME>/<NAMESPACE>/<PROVIDER NAME>/<VERSION>` as the import ID.

For example a private provider version:

```shell
terraform import tfe_registry_provider_version.example my-org-name/private/my-org-name/my-provider/1.0.0
```

Or a public provider version:

```shell
terraform import tfe_registry_provider_version.example my-org-name/public/hashicorp/aws/5.0.0
