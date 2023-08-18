---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_variable_set"
description: |-
  Add a variable set to a workspace
---

# tfe_workspace_variable_set

Adds and removes variable sets from a workspace

-> **Note:** `tfe_variable_set` has a deprecated argument `workspace_ids` that should not be used alongside this resource. They attempt to manage the same attachments and are mutually exclusive.

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
  organization  = tfe_organization.test.name
}

resource "tfe_workspace_variable_set" "test" {
  variable_set_id = tfe_variable_set.test.id
  workspace_id    = tfe_workspace.test.id
}
```

## Argument Reference

The following arguments are supported:

* `variable_set_id` - (Required) The variable set ID.
* `workspace_id` - (Required) Workspace ID to add the variable set to.

## Attributes Reference

* `id` - The ID of the variable set attachment. ID format: `<workspace-id>_<variable-set-id>`

## Import

Workspace Variable Sets can be imported; use `<ORGANIZATION>/<WORKSPACE NAME>/<VARIABLE SET NAME>`. For example:

```shell
terraform import tfe_workspace_variable_set.test 'my-org-name/workspace/My Variable Set'
```
