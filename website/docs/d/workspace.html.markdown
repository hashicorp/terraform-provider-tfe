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
  name = "my-workspace-name"
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the workspace.
* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the workspace within Terraform. This is a custom ID that is
  needed because the Terraform Enterprise workspace related API calls require
  the organization and workspace name instead of the actual workspace ID.
* `auto_apply` - Indicated whether to automatically apply changes when a
  Terraform plan is successful.
* `ssh_key_id` - The ID of an SSH key assigned to the workspace.
* `queue_all_runs` - Indicated whether all runs should be queued.
* `terraform_version` - The version of Terraform used for this workspace.
* `working_directory` - A relative path that Terraform will execute within.
* `vcs_repo` - Settings for the workspace's VCS repository.
* `external_id` - The external ID of the workspace within Terraform Enterprise.

The `vcs_repo` block contains:

* `identifier` - A reference to your VCS repository in the format `:org/:repo`
  where `:org` and `:repo` refer to the organization and repository in your VCS
  provider.
* `branch` - The repository branch that Terraform will execute from.
* `ingress_submodules` - Indicated whether submodules should be fetched when
  cloning the VCS repository.
* `oauth_token_id` - OAuth token ID of the configured VCS connection.
