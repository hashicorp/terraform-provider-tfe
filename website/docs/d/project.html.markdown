---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_project"
description: |-
Get information on a Project.
---

# Data Source: tfe_project

Use this data source to get information about a project.

~> **NOTE:** The `workspace_ids` and `workspace_names` attributes are not guaranteed to return values in the same order, so they cannot be reliably mapped to one another. To map workspace names to IDs reliably, it is recommended to pass those names into the `tfe_workspace_ids` data source.

## Example Usage

```hcl
data "tfe_project" "foo" {
  name = "my-project-name"
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:
* `name` - (Required) Name of the project.
* `organization` - (Required) Name of the organization.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The project ID.
* `workspace_ids` - IDs of the workspaces that are associated with the project.
* `workspace_names` - Names of the workspaces that are associated with the project.
