---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_policy_set_parameter"
description: |-
  Manages policy set parameters.
---

# tfe_policy_set_parameter

Creates, updates and destroys policy set parameters.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_policy_set" "test" {
  name         = "my-policy-set-name"
  organization = tfe_organization.test.id
}

resource "tfe_policy_set_parameter" "test" {
  key          = "my_key_name"
  value        = "my_value_name"
  policy_set_id = tfe_policy_set.test.id
}
```

## Argument Reference

The following arguments are supported:

* `key` - (Required) Name of the parameter.
* `value` - (Required) Value of the parameter.
* `sensitive` - (Optional) Whether the value is sensitive. If true then the
  parameter is written once and not visible thereafter. Defaults to `false`.
* `policy_set_id` - (Required) The ID of the policy set that owns the parameter.

## Attributes Reference

* `id` - The ID of the parameter.

## Import

Parameters can be imported; use
`<POLICY SET ID>/<PARAMETER ID>` as the import ID. For
example:

```shell
terraform import tfe_policy_set_parameter.test polset-wAs3zYmWAhYK7peR/var-5rTwnSaRPogw6apb
```

