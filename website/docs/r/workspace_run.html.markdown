---
layout: "tfe"
page_title: "Terraform Enterprise: Resource tfe_workspace_run"
description: |-
  Provides a resource to manage the initial and/or final Terraform run in a given workspace. These initial and final runs often have a special relationship to other things that depend on the workspace's existence, so it can be useful to manage the completion of these runs in the same Terraform configuration that manages the workspace.
  ~> Note: Use caution when removing tfe_workspace_run from configuration. Destroying with a destroy block present creates a destroy run for underlying managed resources.
  There are a few main use cases this resource was designed for:
  Workspaces that depend on other workspaces. If a workspace will create infrastructure that other workspaces rely on (for example, a Kubernetes cluster to deploy resources into), those downstream workspaces can depend on an initial apply with wait_for_run = true, so they aren't created before their infrastructure dependencies.A more reliable queue_all_runs = true. The queue_all_runs argument on tfe_workspace requests an initial run, which can complete asynchronously outside of the Terraform run that creates the workspace. Unfortunately, it can't be used with workspaces that require variables to be set, because the tfe_variable resources themselves depend on the tfe_workspace. By managing an initial apply with wait_for_run = false that depends on your tfe_variables, you can accomplish the same goal without a circular dependency.Safe workspace destruction. To ensure a workspace's managed resources are destroyed before deleting it, add a destroy block with wait_for_run = true. When you destroy the tfe_workspace_run resource, Terraform will wait for the destroy run to complete before deleting the workspace. This pattern is compatible with the tfe_workspace resource's default safe deletion behavior.
  The tfe_workspace_run expects to own exactly one apply during a creation and/or one destroy during a destruction. This implies that even if previous successful applies exist in the workspace, a tfe_workspace_run resource that includes an apply block will queue a new apply when added to a config.
  -> Note: Using manual_confirm will override the workspace's default apply mode. To use the workspace default apply mode, look up the setting for auto_apply with the tfe_workspace data source.
  ~> Note: If a destroy run cannot be created because the workspace has no configuration version (for example, an empty workspace that never had a configuration uploaded), the destroy is automatically treated as a no-op success. This follows the standard Terraform convention of treating the destruction of an already-absent resource as a success.
---

# Resource: tfe_workspace_run

Provides a resource to manage the _initial_ and/or _final_ Terraform run in a given workspace. These initial and final runs often have a special relationship to other things that depend on the workspace's existence, so it can be useful to manage the completion of these runs in the same Terraform configuration that manages the workspace.

~> **Note:** Use caution when removing `tfe_workspace_run` from configuration. Destroying with a `destroy` block present creates a destroy run for underlying managed resources.

There are a few main use cases this resource was designed for: 
 - **Workspaces that depend on other workspaces.** If a workspace will create infrastructure that other workspaces rely on (for example, a Kubernetes cluster to deploy resources into), those downstream workspaces can depend on an initial `apply` with `wait_for_run = true`, so they aren't created before their infrastructure dependencies.
- **A more reliable `queue_all_runs = true`.** The `queue_all_runs` argument on `tfe_workspace` requests an initial run, which can complete asynchronously outside of the Terraform run that creates the workspace. Unfortunately, it can't be used with workspaces that require variables to be set, because the `tfe_variable` resources themselves depend on the `tfe_workspace`. By managing an initial `apply` with `wait_for_run = false` that depends on your `tfe_variables`, you can accomplish the same goal without a circular dependency.
- **Safe workspace destruction.** To ensure a workspace's managed resources are destroyed before deleting it, add a `destroy` block with `wait_for_run = true`. When you destroy the `tfe_workspace_run` resource, Terraform will wait for the destroy run to complete before deleting the workspace. This pattern is compatible with the `tfe_workspace` resource's default safe deletion behavior.
The `tfe_workspace_run` expects to own exactly one apply during a creation and/or one destroy during a destruction. This implies that even if previous successful applies exist in the workspace, a `tfe_workspace_run` resource that includes an `apply` block will queue a new apply when added to a config.

-> **Note:** Using `manual_confirm` will override the workspace's default apply mode. To use the workspace default apply mode, look up the setting for `auto_apply` with the `tfe_workspace` data source.

~> **Note:** If a `destroy` run cannot be created because the workspace has no configuration version (for example, an empty workspace that never had a configuration uploaded), the destroy is automatically treated as a no-op success. This follows the standard Terraform convention of treating the destruction of an already-absent resource as a success.

## Example Usage

```terraform
# Basic usage with multiple workspaces

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@example.com"
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.test-organization.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "oauth_token_id"
  service_provider = "github"
}

resource "tfe_workspace" "parent" {
  name           = "parent-ws"
  organization   = tfe_organization.test-organization.name
  queue_all_runs = false
  vcs_repo {
    branch         = "main"
    identifier     = "my-org-name/vcs-repository"
    oauth_token_id = tfe_oauth_client.test.oauth_token_id
  }
}

resource "tfe_workspace" "child" {
  name           = "child-ws"
  organization   = tfe_organization.test-organization.name
  queue_all_runs = false
  vcs_repo {
    branch         = "main"
    identifier     = "my-org-name/vcs-repository"
    oauth_token_id = tfe_oauth_client.test.oauth_token_id
  }
}

resource "tfe_workspace_run" "ws_run_parent" {
  workspace_id = tfe_workspace.parent.id

  apply {
    manual_confirm    = false
    wait_for_run      = true
    retry_attempts    = 5
    retry_backoff_min = 5
  }

  destroy {
    manual_confirm    = false
    wait_for_run      = true
    retry_attempts    = 3
    retry_backoff_min = 10
  }
}

resource "tfe_workspace_run" "ws_run_child" {
  workspace_id = tfe_workspace.child.id
  depends_on   = [tfe_workspace_run.ws_run_parent]

  apply {
    manual_confirm    = false
    retry_attempts    = 5
    retry_backoff_min = 5
  }

  destroy {
    manual_confirm    = false
    wait_for_run      = true
    retry_attempts    = 3
    retry_backoff_min = 10
  }
}
```

```terraform
# With manual confirmation

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@example.com"
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.test-organization.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "oauth_token_id"
  service_provider = "github"
}

resource "tfe_workspace" "parent" {
  name           = "parent-ws"
  organization   = tfe_organization.test-organization.name
  queue_all_runs = false
  vcs_repo {
    branch         = "main"
    identifier     = "my-org-name/vcs-repository"
    oauth_token_id = tfe_oauth_client.test.oauth_token_id
  }
}

resource "tfe_workspace_run" "ws_run_parent" {
  workspace_id = tfe_workspace.parent.id

  apply {
    manual_confirm = true
    message        = "test message"
  }

  destroy {
    manual_confirm = true
    wait_for_run   = true
  }
}
```

```terraform
# With no retries

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@example.com"
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.test-organization.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "oauth_token_id"
  service_provider = "github"
}

resource "tfe_workspace" "parent" {
  name           = "parent-ws"
  organization   = tfe_organization.test-organization.name
  queue_all_runs = false
  vcs_repo {
    branch         = "main"
    identifier     = "my-org-name/vcs-repository"
    oauth_token_id = tfe_oauth_client.test.oauth_token_id
  }
}

resource "tfe_workspace_run" "ws_run_parent" {
  workspace_id = tfe_workspace.parent.id

  apply {
    manual_confirm = false
    retry          = false
  }

  destroy {
    manual_confirm = false
    retry          = false
    wait_for_run   = true
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `workspace_id` (String) ID of the workspace to execute the run.

### Optional

- `apply` (Block List, Max: 1) Adding an apply block ensures an apply run is queued when the resource is created. The block controls settings for the workspace's apply run during creation. (see [below for nested schema](#nestedblock--apply))
- `destroy` (Block List, Max: 1) Adding a destroy block ensures a destroy run is queued when the resource is destroyed. The block controls settings for the workspace's destroy run during destruction. (see [below for nested schema](#nestedblock--destroy))

### Read-Only

- `id` (String) The ID of the run created by this resource.

<a id="nestedblock--apply"></a>
### Nested Schema for `apply`

Required:

- `manual_confirm` (Boolean) If set to `true` a human will have to manually confirm a plan in HCP Terraform's UI to start an apply. If set to `false`, this resource will be automatically applied. Defaults to `false`. If `wait_for_run` is set to `false`, this auto-apply will be done by HCP Terraform. If `wait_for_run` is set to `true`, the apply will be confirmed by the provider. The exception is the case of policy check soft-failed where a human has to perform an override by manually confirming the plan even though `manual_confirm` is set to false.

Optional:

- `message` (String) A custom message to associate with the run. If omitted, the default run message is used. Defaults to `Triggered by tfe_workspace_run resource via terraform-provider-tfe on <date>`.
- `retry` (Boolean) Whether or not to retry on plan or apply errors. When set to `true`, `retry_attempts` must also be greater than zero in order for retries to happen. Defaults to `true`.
- `retry_attempts` (Number) The number of retry attempts made after an initial error. Defaults to `3`.
- `retry_backoff_max` (Number) The maximum time in seconds to backoff before attempting a retry. Defaults to `30`.
- `retry_backoff_min` (Number) The minimum time in seconds to backoff before attempting a retry. Defaults to `1`.
- `wait_for_run` (Boolean) Whether or not to wait for a run to reach completion before considering this a success. When set to `false`, the provider considers the `tfe_workspace_run` resource to have been created immediately after the run has been queued. When set to `true`, the provider waits for a successful apply on the target workspace (or a no-change plan). Defaults to `true`.


<a id="nestedblock--destroy"></a>
### Nested Schema for `destroy`

Required:

- `manual_confirm` (Boolean) If set to `true` a human will have to manually confirm a plan in HCP Terraform's UI to start an apply. If set to `false`, this resource will be automatically applied. Defaults to `false`. If `wait_for_run` is set to `false`, this auto-apply will be done by HCP Terraform. If `wait_for_run` is set to `true`, the apply will be confirmed by the provider. The exception is the case of policy check soft-failed where a human has to perform an override by manually confirming the plan even though `manual_confirm` is set to false.

Optional:

- `message` (String) A custom message to associate with the run. If omitted, the default run message is used. Defaults to `Triggered by tfe_workspace_run resource via terraform-provider-tfe on <date>`.
- `retry` (Boolean) Whether or not to retry on plan or apply errors. When set to `true`, `retry_attempts` must also be greater than zero in order for retries to happen. Defaults to `true`.
- `retry_attempts` (Number) The number of retry attempts made after an initial error. Defaults to `3`.
- `retry_backoff_max` (Number) The maximum time in seconds to backoff before attempting a retry. Defaults to `30`.
- `retry_backoff_min` (Number) The minimum time in seconds to backoff before attempting a retry. Defaults to `1`.
- `wait_for_run` (Boolean) Whether or not to wait for a run to reach completion before considering this a success. When set to `false`, the provider considers the `tfe_workspace_run` resource to have been created immediately after the run has been queued. When set to `true`, the provider waits for a successful apply on the target workspace (or a no-change plan). Defaults to `true`.



