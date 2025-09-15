---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_sentinel_version"
description: |-
  Manages Sentinel versions
---

# tfe_sentinel_version

Manage Sentinel versions available on HCP Terraform and Terraform Enterprise.

## Example Usage

Basic Usage:

```hcl
resource "tfe_sentinel_version" "test" {
  version = "0.24.0-custom"
  url = "https://tfe-host.com/path/to/sentinel.zip"
  sha = "e75ac73deb69a6b3aa667cb0b8b731aee79e2904"
}
```

## Argument Reference

The following arguments are supported:

* `version` - (Required) A semantic version string in N.N.N or N.N.N-bundleName format.
* `url` - (Soon to be deprecated) The URL where a ZIP-compressed 64-bit Linux binary of this version can be downloaded.
* `sha` - (Soon to be deprecated) The SHA-256 checksum of the compressed Sentinel binary.
* `official` - (Optional) Whether or not this is an official release of Sentinel. Defaults to "false".
* `enabled` - (Optional) Whether or not this version of Sentinel is enabled for use in HCP Terraform and Terraform Enterprise. Defaults to "true".
* `beta` - (Optional) Whether or not this version of Sentinel is beta pre-release. Defaults to "false".
* `deprecated` - (Optional) Whether or not this version of Sentinel is deprecated. Defaults to "false".
* `deprecated_reason` - (Optional) Additional context about why a version of Sentinel is deprecated. Defaults to "null" unless `deprecated` is true.
* `archs` - (Optional) A list of architecture-specific binaries for this Terraform version. Each entry in the list is a map containing the following attributes:
    * `url` - (Required) The URL where a ZIP-compressed binary of this version can be downloaded.
    * `sha` - (Required) The SHA-256 checksum of the compressed binary.
    * `os` - (Required) The operating system for which this binary is intended.
    * `arch` - (Required) The architecture for which this binary is intended.

    When specifying architecture-specific binaries, the top-level `url` and `sha` attributes are deprecated and should not be used. If both top-level `url` and `sha` are specified, an `archs` entry for the `amd64` architecture must also be included, and its `url` and `sha` values must match the top-level values.

## Attributes Reference

* `id` The ID of the Sentinel version

## Import

Sentinel versions can be imported; use `<SENTINEL VERSION ID>` or `<SENTINEL VERSION NUMBER>` as the import ID. For example:

```shell
terraform import tfe_sentinel_version.test tool-L4oe7rNwn7J4E5Yr
```

```shell
terraform import tfe_sentinel_version.test 0.24.0
```

-> **Note:** You can fetch a Sentinel version ID from the URL of an existing version in the HCP Terraform UI. The ID is in the format `tool-<RANDOM STRING>`
