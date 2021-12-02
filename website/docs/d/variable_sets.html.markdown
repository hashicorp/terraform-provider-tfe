---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_variable_set"
sidebar_current: "docs-datasource-tfe-variable-set"
description: |-
  Get information on organization variable sets.
---

# Data Source: tfe_variable_set

This data source is used to retrieve all variables defined in a specified workspace

## Example Usage

For workspace variables:

```hcl
data "tfe_variable_set" "test" {
  name         = "my-variable-set-name"
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the workspace.
* `organization` - (Required) Name of the organization.

## Attributes Reference

* `id` - The ID of the variable.
* `organization` - Name of the organization.
* `name` - Name of the variable set.
* `description` - Description of the variable set.
* `global` - Whether or not the variable set applies to all workspaces in the organization.
* `workspaces` - IDs of the workspaces that use the variable set.
* `variables` - IDs of the variables attached to the variable set.
