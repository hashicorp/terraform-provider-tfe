---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace"
sidebar_current: "docs-resource-tfe-workspace"
description: |-
  Manages workspaces.
---

# tfe_workspace

Provides a workspace resource.

~> **NOTE:** Using `global_remote_state` or `remote_state_consumer_ids` requires using the provider with Terraform Cloud or an instance of Terraform Enterprise at least as recent as v202104-1.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.id
}
```

(**TFC only**) With `execution_mode` of `agent`:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "test-agent-pool" {
  name         = "my-agent-pool-name"
  organization = tfe_organization.test-organization.id
}

resource "tfe_workspace" "test" {
  name           = "my-workspace-name"
  organization   = tfe_organization.test-organization.id
  agent_pool_id  = tfe_organization.test-agent-pool.id
  execution_mode = "agent"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the workspace.
* `organization` - (Required) Name of the organization.
* `description` - (Optional) A description for the workspace.
* `agent_pool_id` - (Optional) The ID of an agent pool to assign to the workspace. Requires `execution_mode`
  to be set to `agent`. This value _must not_ be provided if `execution_mode` is set to any other value or if `operations` is
  provided.
* `allow_destroy_plan` - (Optional) Whether destroy plans can be queued on the workspace.
* `auto_apply` - (Optional) Whether to automatically apply changes when a
  Terraform plan is successful. Defaults to `false`.
* `execution_mode` - (Optional) Which [execution mode](https://www.terraform.io/docs/cloud/workspaces/settings.html#execution-mode) to use. Using Terraform Cloud, valid
  values are `remote`, `local` or `agent`. Using Terraform Enterprise, only `remote` and `local` execution modes are
  valid.  When set to `local`, the workspace will be used for state storage only. Defaults to `remote`. This value _must
  not_ be provided if `operations` is provided.
* `file_triggers_enabled` - (Optional) Whether to filter runs based on the changed files 
  in a VCS push. If enabled, the working directory and trigger prefixes describe a set of 
  paths which must contain changes for a VCS push to trigger a run. If disabled, any push will 
  trigger a run. Defaults to `true`.
* `global_remote_state` - (Optional) Whether the workspace allows all workspaces in the organization to access its state data during runs. If false, then only specifically approved workspaces can access its state (`remote_state_consumer_ids`).
* `remote_state_consumer_ids` - (Optional) The set of workspace IDs set as explicit remote state consumers for the given workspace.
* `operations` - **Deprecated** Whether to use remote execution mode. When set to `false`, the workspace will 
  be used for state storage only. Defaults to `true`. This value _must not_ be provided if `execution_mode` is 
  provided.
* `queue_all_runs` - (Optional) Whether the workspace should start automatically performing
  runs immediately after its creation. When set to `false`, runs triggered by a webhook
  (such as a commit in VCS) will not be queued until at least one run has been manually
  queued. Defaults to `true`. **Note:** This default differs from the Terraform Cloud API default, which is `false`.
  The provider uses `true` as any workspace provisioned with `false` would need to then have a run manually queued out-of-band
  before accepting webhooks.
* `speculative_enabled` - (Optional) Whether this workspace allows speculative
  plans. Setting this to `false` prevents Terraform Cloud or the Terraform
  Enterprise instance from running plans on pull requests, which can improve
  security if the VCS repository is public or includes untrusted contributors.
  Defaults to `true`.
* `structured_run_output_enabled` - (Optional) Whether this workspace should show output from Terraform runs using the enhanced UI when available. Setting this to `false` ensures that all runs in this workspace will display their output as text logs. Defaults to `true`.
* `ssh_key_id` - (Optional) The ID of an SSH key to assign to the workspace.
* `terraform_version` - (Optional) The version of Terraform to use for this workspace. Defaults to 
  the latest available version.
* `trigger_prefixes` - (Optional) List of repository-root-relative paths which describe all locations 
  to be tracked for changes.
* `working_directory` - (Optional) A relative path that Terraform will execute
  within.  Defaults to the root of your repository.
* `vcs_repo` - (Optional) Settings for the workspace's VCS repository, enabling the [UI/VCS-driven run workflow](https://www.terraform.io/docs/cloud/run/ui.html).
  Omit this argument to utilize the [CLI-driven](https://www.terraform.io/docs/cloud/run/cli.html) and [API-driven](https://www.terraform.io/docs/cloud/run/api.html)
  workflows, where runs are not driven by webhooks on your VCS provider.

The `vcs_repo` block supports:

* `identifier` - (Required) A reference to your VCS repository in the format
  `<organization>/<repository>` where `<organization>` and `<repository>` refer to the organization and repository
  in your VCS provider. The format for Azure DevOps is <organization>/<project>/_git/<repository>.
* `branch` - (Optional) The repository branch that Terraform will execute from.
  This defaults to the repository's default branch (e.g. main).
* `ingress_submodules` - (Optional) Whether submodules should be fetched when
  cloning the VCS repository. Defaults to `false`.
* `oauth_token_id` - (Required) The VCS Connection (OAuth Connection + Token) to use.
  This ID can be obtained from a `tfe_oauth_client` resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The workspace ID.

## Import

~> **NOTE** In versions < 0.15.1, the import ID was in the format `<ORGANIZATION NAME>/<WORKSPACE NAME>`.
This format has been deprecated in favor of the immutable workspace ID in the format `ws-<RANDOM STRING>`.

Workspaces can be imported; use `<WORKSPACE ID>` as the
import ID. For example:

```shell
terraform import tfe_workspace.test ws-CH5in3chf8RJjrVd
```
