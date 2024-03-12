---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_sentinel_version"
description: |-
  Manages Sentinel versions
---

# tfe_sentinel_version

Manage Sentinel versions available on Terraform Cloud/Enterprise.

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
* `url` - (Required) The URL where a ZIP-compressed 64-bit Linux binary of this version can be downloaded.
* `sha` - (Required) The SHA-256 checksum of the compressed Sentinel binary.
* `official` - (Optional) Whether or not this is an official release of Sentinel. Defaults to "false".
* `enabled` - (Optional) Whether or not this version of Sentinel is enabled for use in Terraform Cloud/Enterprise. Defaults to "true".
* `beta` - (Optional) Whether or not this version of Sentinel is beta pre-release. Defaults to "false".
* `deprecated` - (Optional) Whether or not this version of Sentinel is deprecated. Defaults to "false".
* `deprecated_reason` - (Optional) Additional context about why a version of Sentinel is deprecated. Defaults to "null" unless `deprecated` is true.

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

-> **Note:** You can fetch a Sentinel version ID from the URL of an existing version in the Terraform Cloud UI. The ID is in the format `tool-<RANDOM STRING>`
