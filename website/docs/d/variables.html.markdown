---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_variables"
sidebar_current: "docs-datasource-tfe-variables-x"
description: |-
  Get information on a workspace variables.
---

# Data Source: tfe_variables

This data source is used to retrieve all variables defined in a specified workspace

## Example Usage

```hcl
data "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = "my-org-name"
}

data "tfe_variables" "test" {
  workspace_id = data.tfe_workspace.test.id
}
```

## Argument Reference

The following arguments are supported:

* `workspace_id` - (Required) ID of the workspace.

## Attributes Reference

* `variables` - List containing all terraform and environment variables configured on the workspace
* `terraform` - List containing terraform variables configured on the workspace
* `env` - List containing environment variables configured on the workspace

The `variables, terraform and env` blocks contains:

* `id` - The variable Id
* `name` - The variable Key name
* `value` -  The variable value. If the variable is sensitive this value will be empty.
* `category` -  The category of the variable (terraform or environment)
* `sensitive` - If the variable is marked as sensitive or not
* `hcl` - If the variable is marked as HCL or not
