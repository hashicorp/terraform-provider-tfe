---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace"
description: |-
  Get information on a workspace.
---

# Data Source: tfe_workspace

Use this data source to get information about a workspace.

~> **NOTE:** Using `global_remote_state` or `remote_state_consumer_ids` requires using the provider with HCP Terraform or an instance of Terraform Enterprise at least as recent as v202104-1.

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
* `apply_duration_average` - The average duration of applies for this workspace.
* `assessments_enabled` - (Available only in HCP Terraform) Indicates whether health assessments such as drift detection are enabled for the workspace.
* `auto_apply` - Indicates whether to automatically apply changes when a Terraform plan is successful.
* `auto_apply_run_trigger` - Whether the workspace will automatically apply changes for runs that were created by run triggers from another workspace.
* `auto_destroy_activity_duration` - A duration string representing time after workspace activity when an auto-destroy run will be triggered.
* `auto_destroy_at` - Future date/time string at which point all resources in a workspace will be scheduled to be deleted.
* `created_at` - The time when the workspace was created.
* `environment` - The environment of the workspace.
* `file_triggers_enabled` - Indicates whether runs are triggered based on the changed files in a VCS push (if `true`) or always triggered on every push (if `false`).
* `global_remote_state` - (Optional) Whether the workspace should allow all workspaces in the organization to access its state data during runs. If false, then only specifically approved workspaces can access its state (determined by the `remote_state_consumer_ids` argument).
* `hyok_enabled` - (Optional) Whether HYOK is enabled for the workspace.
* `inherits_project_auto_destroy` - Indicates whether this workspace inherits project auto destroy settings.
* `locked` - Indicates whether the workspace is locked.
* `remote_state_consumer_ids` - (Optional) A set of workspace IDs that will be set as the remote state consumers for the given workspace. Cannot be used if `global_remote_state` is set to `true`.
* `operations` - Indicates whether the workspace is using remote execution mode. Set to `false` to switch execution mode to local. `true` by default.
* `plan_duration_average` - The average duration of plans for this workspace.
* `policy_check_failures` - The number of policy check failures from the latest run.
* `project_id` - ID of the workspace's project
* `queue_all_runs` - Indicates whether the workspace will automatically perform runs
  in response to webhooks immediately after its creation. If `false`, an initial run must
  be manually queued to enable future automatic runs.
* `resource_count` - The number of resources managed by the workspace.
* `run_failures` - The number of run failures on the workspace.
* `runs_count` - The number of runs on the workspace.
* `source` - The source of the workspace.
* `source_name` - The name of the workspace creation source, if set.
* `source_url` - The URL of the workspace creation source, if set.
* `speculative_enabled` - Indicates whether this workspace allows speculative plans.
* `ssh_key_id` - The ID of an SSH key assigned to the workspace.
* `structured_run_output_enabled` - Indicates whether runs in this workspace use the enhanced apply UI.
* `effective_tags` - A map of key-value tags associated with the workspace, including any inherited tags from the parent project.
* `tag_names` - The names of tags added to this workspace.
* `terraform_version` - The version (or version constraint) of Terraform used for this workspace.
* `trigger_prefixes` - List of trigger prefixes that describe the paths HCP Terraform monitors for changes, in addition to the working directory. Trigger prefixes are always appended to the root directory of the repository.
  HCP Terraform or Terraform Enterprise will start a run when files are changed in any directory path matching the provided set of prefixes.
* `trigger_patterns` - List of [glob patterns](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/vcs#glob-patterns-for-automatic-run-triggering) that describe the files HCP Terraform monitors for changes. Trigger patterns are always appended to the root directory of the repository.
* `updated_at` - The time when the workspace was last updated.
* `vcs_repo` - Settings for the workspace's VCS repository.
* `working_directory` - A relative path that Terraform will execute within.
* `execution_mode` - Indicates the [execution mode](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode) of the workspace. **Note:** This value might be derived from an organization-level default or set on the workspace itself; see the [`tfe_workspace_settings` resource](tfe_workspace_settings) for details.
* `html_url` - The URL to the browsable HTML overview of the workspace
* `setting_overwrites` - Settings that are overwritten for this workspace.
* `permissions` - The permissions for the current user on this workspace.
* `actions` - Actions that can be performed on this workspace.

The `setting_overwrites` block contains:

* `execution-mode` - Whether execution mode is overwritten at the workspace level.
* `agent-pool` - Whether agent pool is overwritten at the workspace level.

The `permissions` block contains:

* `can-update` - Can update the workspace.
* `can-destroy` - Can destroy the workspace.
* `can-queue-run` - Can queue runs.
* `can-queue-apply` - Can queue apply.
* `can-queue-destroy` - Can queue destroy.
* `can-lock` - Can lock the workspace.
* `can-unlock` - Can unlock the workspace.
* `can-force-unlock` - Can force unlock the workspace.
* `can-read-settings` - Can read workspace settings.
* `can-update-variable` - Can update variables.
* `can-manage-run-tasks` - Can manage run tasks.
* `can-force-delete` - Can force delete the workspace.

The `actions` block contains:

* `is-destroyable` - Whether the workspace can be destroyed.

The `vcs_repo` block contains:

* `identifier` - A reference to your VCS repository in the format `<vcs organization>/<repository>`
  where `<vcs organization>` and `<repository>` refer to the organization and repository in your VCS
  provider.
* `branch` - The repository branch that Terraform will execute from.
* `ingress_submodules` - Indicates whether submodules should be fetched when
  cloning the VCS repository.
* `oauth_token_id` - OAuth token ID of the configured VCS connection.
* `tags_regex` - A regular expression used to trigger a Workspace run for matching Git tags.
