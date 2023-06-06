---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_run"
description: |-
  Manages run create and destroy lifecycles in a workspace.
---

# tfe_workspace_run

Provides a resource to manage the _initial_ and/or _final_ Terraform run in a given workspace. These initial and final runs often have a special relationship to other things that depend on the workspace's existence, so it can be useful to manage the completion of these runs in the same Terraform configuration that manages the workspace.

There are a few main use cases this resource was designed for:

- **Workspaces that depend on other workspaces.** If a workspace will create infrastructure that other workspaces rely on (for example, a Kubernetes cluster to deploy resources into), those downstream workspaces can depend on an initial `apply` with `wait_for_run = true`, so they aren't created before their infrastructure dependencies.
- **A more reliable `queue_all_runs = true`.** The `queue_all_runs` argument on `tfe_workspace` requests an initial run, which can complete asynchronously outside of the Terraform run that creates the workspace. Unfortunately, it can't be used with workspaces that require variables to be set, because the `tfe_variable` resources themselves depend on the `tfe_workspace`. By managing an initial `apply` with `wait_for_run = false` that depends on your `tfe_variables`, you can accomplish the same goal without a circular dependency.
- **Safe workspace destruction.** To ensure a workspace's managed resources are destroyed before deleting it, manage a `destroy` with `wait_for_run = true`. When you destroy the whole configuration, Terraform will wait for the destroy run to complete before deleting the workspace. This pattern is compatible with the `tfe_workspace` resource's default safe deletion behavior.

The `tfe_workspace_run` expects to own exactly one apply during a creation and/or one destroy during a destruction. This implies that even if previous successful applies exist in the workspace, a `tfe_workspace_run` resource that includes an `apply` block will queue a new apply when added to a config.

## Example Usage

Basic usage with multiple workspaces:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.test-organization
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "oauth_token_id"
  service_provider = "github"
}

resource "tfe_workspace" "parent" {
  name                 = "parent-ws"
  organization         = tfe_organization.test-organization
  queue_all_runs       = false
  vcs_repo {
    branch             = "main"
    identifier         = "my-org-name/vcs-repository"
    oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  }
}

resource "tfe_workspace" "child" {
  name                 = "child-ws"
  organization         = tfe_organization.test-organization
  queue_all_runs       = false
  vcs_repo {
    branch             = "main"
    identifier         = "my-org-name/vcs-repository"
    oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  }
}

resource "tfe_workspace_run" "ws_run_parent" {
  workspace_id    = tfe_workspace.parent.id

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
  workspace_id    = tfe_workspace.child.id
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

With manual confirmation:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.test-organization
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "oauth_token_id"
  service_provider = "github"
}

resource "tfe_workspace" "parent" {
  name                 = "parent-ws"
  organization         = tfe_organization.test-organization
  queue_all_runs       = false
  vcs_repo {
    branch             = "main"
    identifier         = "my-org-name/vcs-repository"
    oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  }
}

resource "tfe_workspace_run" "ws_run_parent" {
  workspace_id     = tfe_workspace.parent.id

  apply {
    manual_confirm = true
  }

  destroy {
    manual_confirm = true
    wait_for_run   = true
  }
}

```

With no retries:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.test-organization
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "oauth_token_id"
  service_provider = "github"
}

resource "tfe_workspace" "parent" {
  name                 = "parent-ws"
  organization         = tfe_organization.test-organization
  queue_all_runs       = false
  vcs_repo {
    branch             = "main"
    identifier         = "my-org-name/vcs-repository"
    oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  }
}

resource "tfe_workspace_run" "ws_run_parent" {
  workspace_id    = tfe_workspace.parent.id

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

## Argument Reference

The following arguments are supported:

* `workspace_id` - (Required) ID of the workspace to execute the run.
* `apply` - (Optional) Settings for the workspace's apply run during creation.
* `destroy` - (Optional) Settings for the workspace's destroy run during destruction.

Both `apply` and `destroy` block supports:

* `manual_confirm` - (Required) If set to true a human will have to manually confirm a plan in Terraform Cloud's UI to start an apply. If set to false, this resource will be automatically applied. Defaults to `false`.
  * If `wait_for_run` is set to `false`, this auto-apply will be done by Terraform Cloud.
  * If `wait_for_run` is set to `true`, the apply will be confirmed by the provider. The exception is the case of policy check soft-failed where a human has to perform an override by manually confirming the plan even though `manual_confirm` is set to false.
  * Note that this setting will override the workspace's default apply mode. To use the workspace default apply mode, look up the setting for `auto_apply` with the `tfe_workspace` data source.
* `retry` - (Optional) Whether or not to retry on plan or apply errors. When set to true, `retry_attempts` must also be greater than zero inorder for retries to happen. Defaults to `true`.
* `retry_attempts` - (Optional) The number to retry attempts made after an initial error. Defaults to `3`.
* `retry_backoff_min` - (Optional) The minimum time in seconds to backoff before attempting a retry. Defaults to `1`.
* `retry_backoff_max` - (Optional) The maximum time in seconds to backoff before attempting a retry. Defaults to `30`.
* `wait_for_run` - (Optional) Whether or not to wait for a run to reach completion before considering this a success. When set to `false`, the provider considers the `tfe_workspace_run` resource to have been created immediately after the run has been queued. When set to `true`, the provider waits for a successful apply on the target workspace to have applied successfully (or if it resulted in a no-change plan). Defaults to `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the run created by this resource. Note, if the resource was created without an `apply{}` configuration block, then this ID will not refer to a real run in Terraform Cloud.