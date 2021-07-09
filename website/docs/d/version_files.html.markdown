---
layout: "tfe"
page_title: "Terraform Enterprise: tfe__version_files"
sidebar_current: "docs-datasource-tfe-version-files"
description: |-
  Manages version files.
---
# Data Source: tfe_version_files

Use this data source to point to a source path that contains files, and
auto generate a checksum of the contents of that directory.

## Example Usage

Pointing to a local directory to upload the sentinel config and policies.

```hcl

data "tfe_version_files" "test" {
  source_path = "policies/my-policy-set"
}
```

## Argument Reference

The following arguments are supported:

* `source_path` - (Required) The path to the directory where the files are located.

## Attributes Reference

* `source_path` - The path to the directory where the files are located.
* `checksum` - The checksum generated from hashing the contents of the files
in the directory of the `source_path`.
