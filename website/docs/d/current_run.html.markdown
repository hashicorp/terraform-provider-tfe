---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_current_run"
sidebar_current: "docs-datasource-tfe-current-run"
description: |-
  Get information on the current run.
---

# Data Source: tfe_current_run

Use this data source to get information about the current run.

## Example Usage

```hcl
data "tfe_current_run" "test" {
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

The following attributes are exported:

* `id` - The run ID.
* `workspace` - A `workspace` block as defined below.

The `workspace` block contains:

* `id` - The workspace ID.
* `name` - The name of the workspace.
* `vcs_repo` - Settings for the workspace's VCS repository.

The `vcs_repo` block contains:

* `identifier` - A reference to your VCS repository in the format `:org/:repo`
  where `:org` and `:repo` refer to the organization and repository in your VCS
  provider.
* `oauth_token_id` - OAuth token ID of the configured VCS connection.
