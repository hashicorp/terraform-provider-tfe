---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_ids"
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

data "tfe_workspace_ids" "dev_env_tags_only" {
  organization = "my-org-name"
  tag_filters {
      include = {
        environment = "dev"
      }
  }
}

data "tfe_workspace_ids" "include_and_exclude" {
  organization = "my-org-name"
  tag_filters {
      include = {
          region = "us-east-1"
      }

      exclude = {
        team = "prodsec"
      }
  }
}

data "tfe_workspace_ids" "exclude_all_matching_key" {
  organization = "my-org-name"
  tag_filters {
      exclude = {
        bad_key = "*"
      }
  }
}

data "tfe_workspace_ids" "prod-apps" {
  tag_names    = ["prod", "app", "aws"]
  organization = "my-org-name"
}

data "tfe_workspace_ids" "prod-only" {
  tag_names    = ["prod"]
  exclude_tags = ["app"]
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported. At least one of `names` or `tag_names` must be present. Both can be used together.

* `names` - (Optional) A list of workspace names to search for. Names that don't
  match a valid workspace will be omitted from the results, but are not an error.

    To select _all_ workspaces for an organization, provide a list with a single
    asterisk, like `["*"]`. The asterisk also supports partial matching on prefix and/or suffix, like `[*-prod]`, `[test-*]`, `[*dev*]`.
* `tag_filters` - (Optional) A set of key-value tag filters to search for workspaces.
* `tag_names` - (Optional) **Deprecated** A list of tag names to search for.
* `exclude_tags` - (Optional) **Deprecated** A list of tag names to exclude when searching.
* `organization` - (Required) Name of the organization.

The `tag_filters` block supports:

* `include`: (Optional) A map of key-value tags the workspaces must contain. Each tag included here will be combined using a logical AND when filtering results.
* `exclude`: (Optional) A map of key-value tags to exclude workspaces from the returned list. To exclude all workspaces containing a specific key, use `"*"` as the value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `full_names` - A map of workspace names and their full names, which look like `<ORGANIZATION>/<WORKSPACE>`.
* `ids` - A map of workspace names and their opaque, immutable IDs, which look like `ws-<RANDOM STRING>`.
