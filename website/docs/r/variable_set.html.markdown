---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_variable_set"
sidebar_current: "docs-resource-tfe-variable-set"
description: |-
  Manages variable sets.
---

# tfe_variable_set

Creates, updates and destroys variable sets.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.name
}

resource "tfe_variable_set" "test" {
  name          = "Test Varset"
  description   = "Some description."
  global        = false
  organization  = tfe_organization.test.name
  workspace_ids = [tfe_workspace.test.id]
}

resource "tfe_variable" "test" {
  key             = "seperate_variable"
  value           = "my_value_name"
  category        = "terraform"
  description     = "a useful description"
  variable_set_id = tfe_variable_set.test.id
}

resource "tfe_variable" "test" {
  key             = "another_variable"
  value           = "my_value_name"
  category        = "env"
  description     = "an environment variable"
  variable_set_id = tfe_variable_set.test.id
}
```

Creating a global variable set:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_variable_set" "test" {
  name         = "Global Varset"
  description  = "Variable set applied to all workspaces."
  global       = true
  organization = tfe_organization.test.name
}

resource "tfe_variable" "test" {
  key             = "seperate_variable"
  value           = "my_value_name"
  category        = "terraform"
  description     = "a useful description"
  variable_set_id = tfe_variable_set.test.id
}

resource "tfe_variable" "test" {
  key             = "another_variable"
  value           = "my_value_name"
  category        = "env"
  description     = "an environment variable"
  variable_set_id = tfe_variable_set.test.id
}
```

Creating a variable set that is applied based on workspace tags:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

data "tfe_workspace_ids" "prod-apps" {
  tag_names    = ["prod", "app", "aws"]
  organization = tfe_organization.test.name
}

resource "tfe_variable_set" "test" {
  name          = "Tag Based Varset"
  description   = "Variable set applied to workspaces based on tag."
  organization  = tfe_organization.test.name
  workspace_ids = tfe_workspace_ids.prod-apps.ids
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the variable set.
* `description` - (Optional) Description of the variable set.
* `global` - (Optional) Whether or not the variable set applies to all workspaces in the organization. Defaults to `false`.
* `organization` - (Required) Name of the organization.
* `workspace_ids` - (Optional) IDs of the workspaces that use the variable set.

## Attributes Reference

* `id` - The ID of the variable set.

## Import

Variable sets can be imported; use `<VARIABLE SET ID>` as the import ID. For example:

```shell
terraform import tfe_variable_set.test varset-5rTwnSaRPogw6apb
```
