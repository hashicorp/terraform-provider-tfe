---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_slug"
sidebar_current: "docs-datasource-tfe-slug"
description: |-
  Manages files.
---
# Data Source: tfe_slug

Use this data source to point to a source path that contains files, and
auto generate a checksum of the contents of that directory. This is used
to determine uniqueness of a directory.

## Example Usage

Pointing to a local directory to upload the sentinel config and policies.

```hcl

data "tfe_slug" "test" {
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
