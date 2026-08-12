---
layout: "tfe"
page_title: "Terraform Enterprise: Resource tfe_variable"
description: |-
  Creates, updates and destroys variables.
  -> Note: While the value field may be referenced in other resources, for safety it is always treated as sensitive. This means that it will always be redacted from plan outputs, and any other resource attributes which depend on it will also be redacted. The readable_value attribute is not sensitive, and will not be redacted; instead, it will be null if the variable is sensitive. This allows other resources to reference it, while keeping their plan outputs readable.
  ~> Note: When sensitive is set to true, Terraform cannot detect and repair drift if value is later changed out-of-band via the HCP Terraform UI. Terraform will only change the value for a sensitive variable if you change value in the configuration, so that it no longer matches the last known value in the state.
---

# Resource: tfe_variable

Creates, updates and destroys variables.

-> **Note:** While the `value` field may be referenced in other resources, for safety it is always treated as sensitive. This means that it will always be redacted from plan outputs, and any other resource attributes which depend on it will also be redacted. The `readable_value` attribute is not sensitive, and will not be redacted; instead, it will be null if the variable is sensitive. This allows other resources to reference it, while keeping their plan outputs readable.

~> **Note:** When `sensitive` is set to `true`, Terraform cannot detect and repair drift if `value` is later changed out-of-band via the HCP Terraform UI. Terraform will only change the value for a sensitive variable if you change `value` in the configuration, so that it no longer matches the last known value in the state.

## Example Usage

```terraform
# Basic usage for workspaces

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

```terraform
# Basic usage for the write-only value of tfe_variable

variable "session_token" {
  type      = string
  ephemeral = true
}

resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.name
}

resource "tfe_variable" "test" {
  key              = "my_key_name"
  value_wo         = var.session_token
  value_wo_version = 1
  category         = "terraform"
  workspace_id     = tfe_workspace.test.id
  description      = "a useful description"
}
```

```terraform
# Basic usage for variable sets

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

```terraform
# Using readable_value

# While the `value` field may be referenced in other resources, for safety it is always treated as sensitive. This means that it will always be redacted from plan outputs, and any other resource attributes which depend on it will also be redacted.

# The `readable_value` attribute is not sensitive, and will not be redacted; instead, it will be null if the variable is sensitive. This allows other resources to reference it, while keeping their plan outputs readable.

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

resource "tfe_workspace" "workspace" {
  name         = "my-workspace"
  organization = "organization name"
}

resource "tfe_workspace" "sensitive_workspace" {
  name         = "workspace-${tfe_variable.sensitive_var.value}" // this will be redacted from plan outputs
  organization = "organization name"
}

resource "tfe_workspace" "visible_workspace" {
  name         = "workspace-${tfe_variable.visible_var.readable_value}" // this will not be redacted from plan outputs
  organization = "organization name"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `category` (String) Whether this is a Terraform or environment variable. Valid values are `terraform` or `env`.
- `key` (String) Name of the variable.

### Optional

> **NOTE**: [Write-only arguments](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments) are supported in Terraform 1.11 and later.

- `description` (String) Description of the variable.
- `hcl` (Boolean) Whether to evaluate the value of the variable as a string of HCL code. Has no effect for environment variables. Defaults to `false`.
- `sensitive` (Boolean) Whether the value is sensitive. If true then the variable is written once and not visible thereafter. Defaults to false.
- `value` (String, Sensitive) Value of the variable. Either `value` or `value_wo` can be provided, but not both.
- `value_wo` (String, Sensitive, [Write-only](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments)) Value of the variable in write-only mode. `Write-only` attributes function similarly to their non-write-only counterparts, but are never stored to state and do not display in the Terraform plan output. Can be used in place of `value`. Either `value` or `value_wo` can be provided, but not both.
- `value_wo_version` (Number) Version identifier for the write-only value. Required when `value_wo` is specified to trigger updates. Cannot be used with `value`.
- `variable_set_id` (String) ID of the variable set that owns the variable. Exactly one of `workspace_id` or `variable_set_id` must be provided.
- `workspace_id` (String) ID of the workspace that owns the variable. Exactly one of `workspace_id` or `variable_set_id` must be provided.

### Read-Only

- `id` (String) The ID of the variable.
- `readable_value` (String) Only present if the variable is non-sensitive. A copy of the value which will not be marked as sensitive in plan outputs. Will be `null` if the variable is sensitive. Cannot be explicitly set in the resource configuration.



## Import

tfe_variable can be imported using an identity. For example:

```terraform
import {
  to = tfe_variable.foo
  identity = {
    id              = "var-5rTwnSaRPogw6apb"
    configurable_id = "ws-66fE3LmF42piTaN2"
    hostname        = "app.terraform.io"
  }
}

import {
  to = tfe_variable.bar
  identity = {
    id              = "var-5rTwnSaRPogw6apb"
    configurable_id = "varset-47qC3LmA47piVan7"
    hostname        = "app.terraform.io"
  }
}
```

<!-- schema generated by tfplugindocs -->
### Identity Schema

#### Required

- `configurable_id` (String)
- `id` (String)

#### Optional

- `hostname` (String)


Resource tfe_variable can be imported in the following format: 

```shell
# via <ORGANIZATION NAME>/<WORKSPACE NAME>/<VARIABLE ID>
terraform import tfe_variable.test my-org-name/my-workspace-name/var-5rTwnSaRPogw6apb

# via <ORGANIZATION NAME>/<VARIABLE SET ID>/<VARIABLE ID>
terraform import tfe_variable.test my-org-name/varset-47qC3LmA47piVan7/var-5rTwnSaRPogw6apb
```
