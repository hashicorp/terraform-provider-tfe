---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_opa_version"
description: |-
  Manages OPA versions
---

# tfe_opa_version

Manage OPA versions available on Terraform Enterprise.

## Example Usage

Basic Usage:

```hcl
resource "tfe_opa_version" "test" {
  version = "0.58.0-custom"
  url = "https://tfe-host.com/path/to/opa"
  sha = "e75ac73deb69a6b3aa667cb0b8b731aee79e2904"
}
```

## Argument Reference

The following arguments are supported:

* `version` - (Required) A semantic version string in N.N.N or N.N.N-bundleName format.
* `url` - (Required) The URL where a 64-bit Linux binary of this version can be downloaded.
* `sha` - (Required) The SHA-256 checksum of the compressed OPA binary.
* `official` - (Optional) Whether or not this is an official release of OPA. Defaults to "false".
* `enabled` - (Optional) Whether or not this version of OPA is enabled for use in HCP Terraform and Terraform Enterprise. Defaults to "true".
* `beta` - (Optional) Whether or not this version of OPA is beta pre-release. Defaults to "false".
* `deprecated` - (Optional) Whether or not this version of OPA is deprecated. Defaults to "false".
* `deprecated_reason` - (Optional) Additional context about why a version of OPA is deprecated. Defaults to "null" unless `deprecated` is true.

## Attributes Reference

* `id` The ID of the OPA version

## Import

OPA versions can be imported; use `<OPA VERSION ID>` or `<OPA VERSION NUMBER>` as the import ID. For example:

```shell
terraform import tfe_opa_version.test tool-L4oe7rNwn7J4E5Yr
```

```shell
terraform import tfe_opa_version.test 0.58.0
```

-> **Note:** You can fetch a OPA version ID from the URL of an existing version in the HCP Terraform UI. The ID is in the format `tool-<RANDOM STRING>`
