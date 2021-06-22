---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_policy_set_version"
sidebar_current: "docs-resource-tfe-tfe-policy-set-version"
description: |-
  Manages policy set versions.
---

# tfe_policy_set_version

A policy set version is a way to version policy sets and upload policies. This resource
enables the ability to upload a set of local policies. 

This resource depends on a data source `tfe_version_files` for managing
the files themselves. 

See example usage for more details.

## Example Usage

Pointing to a local directory to upload the sentinel config and policies.

```hcl

resource "tfe_organization" "test" {
  name  = "<org-name>"
  email = "admin@example.com"
}

resource "tfe_policy_set" "test" {
  name          = "my-policy-set"
  description   = "A brand new policy set"
  organization  = tfe_organization.test.id
}

data "tfe_version_files" "test" {
  source_path = "policies/my-policy-set"
}

resource "tfe_policy_set_version" "test" {
  policy_set_id = tfe_policy_set.test.id
  policies_path_contents_checksum = data.tfe_version_files.test.checksum
  policies_path = data.tfe_version_files.test.source_path
}
```

## Argument Reference

The following arguments are supported:

* `policy_set_id` - (Required) The ID of the Policy Set.
* `policies_path_contents_checksum` - (Required) A checksum generated as a result of hashing
all the contents in the `policies_path`. This is auto generated as a result of using the 
data source `tfe_version_files` field `checksum`. This can be set manually, but that requires
self management of this checksum.
* `policies_path` - (Required) This is the path to the policies. It is highly recommended to use the
data source `tfe_version_files` and field `source_path`. This can also be set manually.

## Attributes Reference

* `id` - The ID of the policy set version.
* `status` - The status of the policy set version.
* `error_message` - The error message for when an error occurs during an
  operation.

