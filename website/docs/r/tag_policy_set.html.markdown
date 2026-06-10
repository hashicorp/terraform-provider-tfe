---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_tag_policy_set"
description: |-
  Add a tag-based inclusion to a policy set
---

# tfe_tag_policy_set

Adds and removes tag-based inclusions on a policy set. Tag inclusions scope policy set enforcement to workspaces that carry a matching tag. If a tag value is not provided, this becomes a key-only tag and only matches workspaces that also have a key-only tag with the given key.

~> **NOTE:** This feature is currently in beta and is not available to all users.

~> **NOTE:** Tag-based scoping and explicit workspace/project associations are mutually exclusive on a policy set. To switch between them, first remove the existing association (`terraform apply`), then add the new one (`terraform apply`). Attempting both in a single apply may fail.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_policy_set" "test" {
  name         = "my-policy-set"
  description  = "Some description."
  organization = tfe_organization.test.name
}

resource "tfe_tag_policy_set" "test" {
  policy_set_id = tfe_policy_set.test.id
  key           = "env"
  value         = "prod"
}
```

Key-only tag (no value):

```hcl
resource "tfe_tag_policy_set" "env_any" {
  policy_set_id = tfe_policy_set.test.id
  key           = "env"
}
```

## Argument Reference

The following arguments are supported:

* `policy_set_id` - (Required) ID of the policy set.
* `key` - (Required) The tag key to match.
* `value` - (Optional) The tag value to match. If omitted, this becomes a key-only tag and only matches workspaces that also have a key-only tag with the given key.

## Import

Tag policy set inclusions can be imported; use `<POLICY_SET_ID>/<TAG_KEY>/<TAG_VALUE>` for key+value tags, or `<POLICY_SET_ID>/<TAG_KEY>` for key-only tags. For example:

```shell
terraform import tfe_tag_policy_set.test 'polset-abc123/env/prod'
```
