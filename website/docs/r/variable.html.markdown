---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_variable"
description: |-
  Manages variables.
---

# tfe_variable

Creates, updates and destroys variables.

## Example Usage

Basic usage for workspaces:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.name
}

resource "tfe_variable" "test" {
  key          = "my_key_name"
  value        = "my_value_name"
  category     = "terraform"
  workspace_id = tfe_workspace.test.id
  description  = "a useful description"
}
```

Basic usage for variable sets:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_variable_set" "test" {
  name         = "Test Varset"
  description  = "Some description."
  global       = false
  organization = tfe_organization.test.name
}

resource "tfe_variable" "test-a" {
  key             = "seperate_variable"
  value           = "my_value_name"
  category        = "terraform"
  description     = "a useful description"
  variable_set_id = tfe_variable_set.test.id
}

resource "tfe_variable" "test-b" {
  key             = "another_variable"
  value           = "my_value_name"
  category        = "env"
  description     = "an environment variable"
  variable_set_id = tfe_variable_set.test.id
}
```

## Argument Reference

The following arguments are supported:

* `key` - (Required) Name of the variable.
* `value` - (Required) Value of the variable.
* `category` - (Required) Whether this is a Terraform or environment variable.
  Valid values are `terraform` or `env`.
* `description` - (Optional) Description of the variable.
* `hcl` - (Optional) Whether to evaluate the value of the variable as a string
  of HCL code. Has no effect for environment variables. Defaults to `false`.
* `sensitive` - (Optional) Whether the value is sensitive. If true then the
variable is written once and not visible thereafter. Defaults to `false`.
* One of the following (Required)
    * `workspace_id` - ID of the workspace that owns the variable.
    * `variable_set_id` - ID of the variable set that owns the variable.

~> **NOTE:** When `sensitive` is set to true, Terraform cannot detect and repair
drift if `value` is later changed out-of-band via the Terraform Cloud UI.
Terraform will only change the value for a sensitive variable if you change
`value` in the configuration, so that it no longer matches the last known value
in the state.

## Attributes Reference

* `id` - The ID of the variable.
* `readable_value` - Only present if the variable is non-sensitive. A copy of the value which will not be marked as sensitive in plan outputs.

### Using readable_value

While the `value` field may be referenced in other resources, for safety it is always treated as sensitive. This means that it will always be redacted from plan outputs, and any other resource attributes which depend on it will also be redacted.

The `readable_value` attribute is not sensitive, and will not be redacted; instead, it will be null if the variable is sensitive. This allows other resources to reference it, while keeping their plan outputs readable.

For example:
```
resource "tfe_variable" "sensitive_var" {
  key          = "sensitive_key"
  value        = "sensitive_value" // this will be redacted from plan outputs
  category     = "terraform"
  workspace_id = tfe_workspace.workspace.id
  sensitive    = true
}

resource "tfe_variable" "visible_var" {
  key          = "visible_key"
  value        = "visible_value" // this will be redacted from plan outputs
  category     = "terraform"
  workspace_id = tfe_workspace.workspace.id
  sensitive    = false
}

resource "tfe_workspace" "sensitive_workspace" {
  name = "workspace-${tfe_variable.sensitive_var.value}" // this will be redacted from plan outputs
  organization = "organization name"
}

resource "tfe_workspace" "visible_workspace" {
  name = "workspace-${tfe_variable.visible_var.readable_value}" // this will not be redacted from plan outputs
  organization = "organization name"
}

```

`readable_value` will be null if the variable is sensitive. `readable_value` may not be set explicitly in the resource configuration.


## Import

Variables can be imported.

To import a variable that's part of a workspace, use
`<ORGANIZATION NAME>/<WORKSPACE NAME>/<VARIABLE ID>` as the import ID. For
example:

```shell
terraform import tfe_variable.test my-org-name/my-workspace-name/var-5rTwnSaRPogw6apb
```

To import a variable that's part of a variable set, use
`<ORGANIZATION NAME>/<VARIABLE SET ID>/<VARIABLE ID>` as the import ID. For
example:

```shell
terraform import tfe_variable.test my-org-name/varset-47qC3LmA47piVan7/var-5rTwnSaRPogw6apb
```
