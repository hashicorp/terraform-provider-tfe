---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_policy_set_version_files"
sidebar_current: "docs-datasource-tfe_policy_set_version_files"
description: |-
  Manages policy set version files.
---
# Data Source: tfe_policy_set_version_files

Use this data source to point to a source path that contains files, and
auto generate a checksum of the contents of that directory.

## Example Usage

Pointing to a local directory to upload the sentinel config and policies.

```hcl

data "tfe_policy_set_version_files" "test" {
  source_path = "policies/my-policy-set"
}
```

## Argument Reference

The following arguments are supported:

* `source_path` - (Required) The path to the directory where the files are located.

## Attributes Reference

* `source_path` - The path to the directory where the files are located.
* `output_sha` - The checksum generated from hashing the contents of the files
in the directory of the `source_path`.
