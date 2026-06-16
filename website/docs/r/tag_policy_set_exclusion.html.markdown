---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_tag_policy_set_exclusion"
description: |-
  Add a tag-based exclusion to a policy set
---

# tfe_tag_policy_set_exclusion

Adds and removes tag-based exclusions on a policy set. Tag exclusions exempt workspaces that carry a matching tag from policy set enforcement. If a tag value is not provided, this becomes a key-only tag and only matches workspaces that also have a key-only tag with the given key.

~> **Note:** `tfe_policy_set` has an argument `global` that should be `true` to use this resource.

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
  global       = true
}

resource "tfe_tag_policy_set_exclusion" "test" {
  policy_set_id = tfe_policy_set.test.id
  key           = "env"
  value         = "staging"
}
```

## Argument Reference

The following arguments are supported:

* `policy_set_id` - (Required) ID of the policy set.
* `key` - (Required) The tag key to match for exclusion.
* `value` - (Optional) The tag value to match. If omitted, this becomes a key-only tag and only matches workspaces that also have a key-only tag with the given key.

## Import

Tag policy set exclusions can be imported; use `<POLICY_SET_ID>/<TAG_KEY>/<TAG_VALUE>` for key+value tags, or `<POLICY_SET_ID>/<TAG_KEY>` for key-only tags. For example:

```shell
terraform import tfe_tag_policy_set_exclusion.test 'polset-abc123/env/staging'
```
