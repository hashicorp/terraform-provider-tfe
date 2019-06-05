---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace"
sidebar_current: "docs-resource-tfe-workspace"
description: |-
  Manages workspaces.
---

# tfe_workspace

Provides a workspace resource.

## Example Usage

Basic usage:

```hcl
resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the workspace.
* `organization` - (Required) Name of the organization.
* `auto_apply` - (Optional) Whether to automatically apply changes when a
  Terraform plan is successful. Defaults to `false`.
* `file_triggers_enabled` - (Optional) Whether to filter runs based on the changed files in a VCS push. If enabled, the working directory and trigger prefixes describe a set of paths which must contain changes for a VCS push to trigger a run. If disabled, any push will trigger a run. Defaults to `true`.
* `queue_all_runs` - (Optional) Whether all runs should be queued. When set
  to `false`, runs triggered by a VCS change will not be queued until at least
  one run is manually queued. Defaults to `true`.
* `ssh_key_id` - (Optional) The ID of an SSH key to assign to the workspace.
* `terraform_version` - (Optional) The version of Terraform to use for this
* `trigger_prefixes` - (Optional) List of repository-root-relative paths which describe all locations to be tracked for changes.
  workspace. Defaults to the latest available version.
* `working_directory` - (Optional) A relative path that Terraform will execute
  within.  Defaults to the root of your repository.
* `vcs_repo` - (Optional) Settings for the workspace's VCS repository.

The `vcs_repo` block supports:

* `identifier` - (Required) A reference to your VCS repository in the format
  `:org/:repo` where `:org` and `:repo` refer to the organization and repository
  in your VCS provider.
* `branch` - (Optional) The repository branch that Terraform will execute from.
  Default to `master`.
* `ingress_submodules` - (Optional) Whether submodules should be fetched when
  cloning the VCS repository. Defaults to `false`.
* `oauth_token_id` - (Required) Token ID of the VCS Connection (OAuth Conection Token)
  to use.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The workspace's human-readable ID, which looks like
  `<ORGANIZATION>/<WORKSPACE>`.
* `external_id` - The workspace's opaque external ID, which looks like
  `ws-<RANDOM STRING>`.

## Import

Workspaces can be imported; use `<ORGANIZATION NAME>/<WORKSPACE NAME>` as the
import ID. For example:

```shell
terraform import tfe_workspace.test my-org-name/my-workspace-name
```
