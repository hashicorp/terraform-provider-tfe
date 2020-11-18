---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_policy_set_version"
sidebar_current: "docs-resource-tfe-policy-set-version"
description: |-
  Manages policy set versions for non-VCS-backed Sentinel policy sets.
---

# tfe_policy_set_version

Creates and updates policy set versions for non-VCS-backed Sentinel policy sets
and loads their files including a `sentinel.hcl` configuration file, Sentinel
policies, and Sentinel modules.

~> Policy Set Versions cannot be directly deleted; they are only deleted when
   the policy set they belong to is deleted. When you update an instance of a `tfe_policy_set_version` resource by changing its `version` or `directory`
   attributes, Terraform will report that the resource is being replaced and
   indicate that one instance is being destroyed and that a new one is being
   created. However, in reality, the provider is only creating a new policy set
   version instance and uploading files to it. Similarly, if you remove an
   instance of the `tfe_policy_set_version` resource from your Terraform
   configuration or run `terraform destroy`, Terraform will report that the
   resource is being destroyed. That is accurate as far as the resource
   itself is concerned; however, the underlying policy set version in your
   Terraform Cloud or Terraform Enterprise organization will not actually be
   deleted.

## Example Usage

Basic usage:

```hcl
resource "tfe_policy_set" "test" {
  name         = "my-policy-set"
  description  = "a brand new policy set"
  organization = "my-org-name"
}

resource "tfe_policy_set_version" "test" {
  version       = "1"
  directory     = "${path.module}/my-policy-set-directory"
  policy_set_id = "${tfe_policy_set.test.id}"
}
```

## Argument Reference

The following arguments are supported:

* `version` - (Required) Name or number of the version. This can be any string.
  Changing this forces creation of a new resource. We suggest using integers or
  version strings with the form `x.y.z` which can be incremented whenever you
  change any of the policy set files.
* `directory` - (Required) The path of the directory containing the Sentinel
  files that should be uploaded.
* `policy_set_id` - (Required) The ID of the policy set that owns the version.

## Attributes Reference

* `id` - The ID of the policy set version.
* `status` - The status of the policy set version. This should be `ready` if
  the policy set files were successfully uploaded. Otherwise, it would be
  `pending`.
* `created_at` - The timestamp when the policy set version was created.
* `updated_at` - The timestamp when the policy set version was last updated.

## Import

Policy set versions can be imported; use
`<POLICY SET ID>/<POLICY SET VERSION ID>` as the import ID. For
example:

```shell
terraform import tfe_policy_set_version.test polset-wAs3zYmWAhYK7peR/polsetver-pAjRi9wUzMqLkGj2
```
