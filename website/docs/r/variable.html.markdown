---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_variable"
sidebar_current: "docs-resource-tfe-variable"
description: |-
  Creates, updates and destroys variables.
---

# tfe_variable

Creates, updates and destroys variables.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "organization" {
  name = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "workspace" {
  name = "my-workspace-name"
  organization = "${tfe_organization.organization.id}"
}

resource "tfe_variable" "variable" {
  key = "my_key_name"
  value = "my_value_name"
  category = "terraform"
  workspace_id = "${tfe_workspace.workspace.id}"
}
```

## Argument Reference

The following arguments are supported:

* `key` - (Required) Name of the variable.
* `value` - (Required) Value of the variable.
* `category` - (Required) Whether this is a Terraform or environment variable.
  Valid values are `terraform` or `env`.
* `hcl` - (Optional) Whether to evaluate the value of the variable as a string
  of HCL code. Has no effect for environment variables. Defaults to `false`.
* `sensitive` - (Optional) Whether the value is sensitive. If true then the
  variable is written once and not visible thereafter. Defaults to `false`.
* `workspace_id` - (Required) ID of the workspace that owns the variable.

## Attributes Reference

* `id` - The ID of the variable.
