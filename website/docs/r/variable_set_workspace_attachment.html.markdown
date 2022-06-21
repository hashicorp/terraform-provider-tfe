---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_variable_set"
sidebar_current: "docs-resource-tfe-variable-set-workspace-attachment"
description: |-
  Add a workspace to a variable set.
---

# tfe_variable_set_workspace_attachment

Creats and destroys workspace attachments to variable sets.

!> **Warning:** `tfe_variable_set` has an argument `workspace_ids` that should not be used alongside this resource. They attempt to manage the same attachments and are mutually exclusive.

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

resource "tfe_variable_set_workspace_attachment" "test" {
  variable_set_id = tfe_variable_set.test.id
  workspace_id    = tfe_workspace.test.id
}
```

## Argument Reference

The following arguments are supported:

* `variable_set_id` - (Required) Name of the variable set.
* `workspace_id` - (Required) Workspace ID to attach to variable set.

## Attributes Reference

* `id` - The ID of the variable set attachment. ID format: `<variable-set-id>_<workspace-id>`

## Import

Variable set workspace attachment can be imported; use `<variable-set-id>_<workspace-id>` as the import ID. For example:

```shell
terraform import tfe_variable_set_workspace_attachment.test 'varset-QDyoQft813kinftv_ws-EnSMN5DkW3KcuYFc'
```
