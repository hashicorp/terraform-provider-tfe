---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_ids"
sidebar_current: "docs-datasource-tfe-workspace-ids"
description: |-
  Get information on (external) workspace IDs.
---

# Data Source: tfe_workspace

Use this data source to get a map of (external) workspace IDs.

## Example Usage

```hcl
data "tfe_workspace" "test" {
  names        = ["my-workspace-name"]
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:
* `names` - (Required) A list of workspace names.
* `organization` - (Required) Name of the organization.

~> The list of names can be used to search for workspaces with matching names.
  Additionally you can also use a single entry with a wildcard (e.g. "*") which
  will match all names. Using a partial string together with a wildcard (e.g.
  "my-workspace-*") is not supported.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ids` - A map of the workspace names and their IDs within Terraform. This is a
  custom ID that is needed because the Terraform Enterprise workspace related
  API calls require the organization and workspace name instead of the actual
  workspace ID.
* `external_ids` - A map of workspace names and their external IDs.
