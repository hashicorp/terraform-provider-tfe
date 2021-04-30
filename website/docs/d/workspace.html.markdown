---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace"
sidebar_current: "docs-datasource-tfe-workspace-x"
description: |-
  Get information on a workspace.
---

# Data Source: tfe_workspace

Use this data source to get information about a workspace.

~> **NOTE:** Using `global_remote_state` or `remote_state_consumer_ids` requires using the provider with Terraform Cloud or an instance of Terraform Enterprise at least as recent as v202104-1.

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

* `id` - The workspace ID.
* `allow_destroy_plan` - Indicates whether destroy plans can be queued on the workspace.
* `auto_apply` - Indicates whether to automatically apply changes when a Terraform plan is successful.
* `file_triggers_enabled` - Indicates whether runs are triggered based on the changed files in a VCS push (if `true`) or always triggered on every push (if `false`).
* `global_remote_state` - (Optional) Whether the workspace should allow all workspaces in the organization to access its state data during runs. If false, then only specifically approved workspaces can access its state (determined by the `remote_state_consumer_ids` argument).
* `remote_state_consumer_ids` - (Optional) A set of workspace IDs that will be set as the remote state consumers for the given workspace. Cannot be used if `global_remote_state` is set to `true`.
* `operations` - Indicates whether the workspace is using remote execution mode. Set to `false` to switch execution mode to local. `true` by default.
* `queue_all_runs` - Indicates whether the workspace will automatically perform runs
  in response to webhooks immediately after its creation. If `false`, an initial run must
  be manually queued to enable future automatic runs.
* `speculative_enabled` - Indicates whether this workspace allows speculative plans.
* `ssh_key_id` - The ID of an SSH key assigned to the workspace.
* `terraform_version` - The version of Terraform used for this workspace.
* `trigger_prefixes` - List of repository-root-relative paths which describe all locations to be tracked for changes.
* `vcs_repo` - Settings for the workspace's VCS repository.
* `working_directory` - A relative path that Terraform will execute within.
* `resource_count` - The number of resources managed by the workspace.
* `policy_check_failures` - The number of policy check failures from the latest run.
* `run_failures` - The number of run failures on the workspace.
* `runs_count` - The number of runs on the workspace.


The `vcs_repo` block contains:

* `identifier` - A reference to your VCS repository in the format `<organization>/<repository>`
  where `<organization>` and `<repository>` refer to the organization and repository in your VCS
  provider.
* `branch` - The repository branch that Terraform will execute from.
* `ingress_submodules` - Indicates whether submodules should be fetched when
  cloning the VCS repository.
* `oauth_token_id` - OAuth token ID of the configured VCS connection.
