---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_slug"
description: |-
  Manages files.
---
# Data Source: tfe_slug

This data source is used to represent configuration files on a local filesystem
intended to be uploaded to Terraform Cloud/Enterprise, in lieu of those files being
sourced from a configured VCS provider.

A unique checksum is generated for the specified local directory, which allows
resources such as `tfe_policy_set` track the files and upload a new gzip compressed
tar file containing configuration files (a Terraform "slug") when those files change.

## Example Usage

Tracking a local directory to upload the Sentinel configuration and policies:

```hcl

data "tfe_slug" "test" {
  source_path = "policies/my-policy-set"
}

resource "tfe_policy_set" "test" {
  name          = "my-policy-set"
  organization  = "my-org-name"

  // reference the tfe_slug data source.
  slug = data.tfe_slug.test
}
```

## Argument Reference

The following arguments are supported:

* `source_path` - (Required) The path to the directory where the files are located.
