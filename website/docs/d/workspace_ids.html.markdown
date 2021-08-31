---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_ids"
sidebar_current: "docs-datasource-tfe-workspace-ids"
description: |-
  Get information on workspace IDs.
---

# Data Source: tfe_workspace_ids

Use this data source to get a map of workspace IDs.

## Example Usage

```hcl
data "tfe_workspace_ids" "app-frontend" {
  names        = ["app-frontend-prod", "app-frontend-dev1", "app-frontend-staging"]
  organization = "my-org-name"
}

data "tfe_workspace_ids" "all" {
  names        = ["*"]
  organization = "my-org-name"
}

data "tfe_workspace_ids" "prod-apps" {
  tags         = ["prod", "app", "aws"]
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `names` - (Required) A list of workspace names to search for. Names that don't
  match a real workspace will be omitted from the results, but are not an error.

    To select _all_ workspaces for an organization, provide a list with a single
    asterisk, like `["*"]`. No other use of wildcards is supported.
* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `full_names` - A map of workspace names and their full names, which look like `<ORGANIZATION>/<WORKSPACE>`. 
* `ids` - A map of workspace names and their opaque, immutable IDs, which look like `ws-<RANDOM STRING>`.
