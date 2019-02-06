---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace"
sidebar_current: "docs-datasource-tfe-workspace-x"
description: |-
  Get information on a workspace.
---

# Data Source: tfe_workspace

Use this data source to get information about a workspace.

## Example Usage

```hcl
data "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the workspace.
* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The workspace's human-readable ID, which looks like
  `<ORGANIZATION>/<WORKSPACE>`.
* `external_id` - The workspace's opaque external ID, which looks like
  `ws-<RANDOM STRING>`.
* `auto_apply` - Indicates whether to automatically apply changes when a
  Terraform plan is successful.
* `queue_all_runs` - Indicates whether all runs should be queued.
* `ssh_key_id` - The ID of an SSH key assigned to the workspace.
* `terraform_version` - The version of Terraform used for this workspace.
* `vcs_repo` - Settings for the workspace's VCS repository.
* `working_directory` - A relative path that Terraform will execute within.

The `vcs_repo` block contains:

* `identifier` - A reference to your VCS repository in the format `:org/:repo`
  where `:org` and `:repo` refer to the organization and repository in your VCS
  provider.
* `ingress_submodules` - Indicates whether submodules should be fetched when
  cloning the VCS repository.
* `oauth_token_id` - OAuth token ID of the configured VCS connection.
