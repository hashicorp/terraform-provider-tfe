---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_run"
description: |-
  Manages run create and destroy lifecycles in a workspace.
---

# tfe_workspace_run

Provides a resource to manage create and destroy lifecycles in a workspace.
The `tfe_workspace_run` expects to own exactly one apply during a creation and/or one destroy during a destruction. This implies that even though previous successful applies exist in the workspace, the `tfe_workspace_run` resource will queue a new apply when added to a config.


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
  organization = tfe_organization.test-organization
  workspace    = tfe_workspace.parent.name

  apply {
    retry_attempts = 5
    retry_backoff_min = 5
  }

  destroy {
    retry_attempts = 3
    retry_backoff_min = 10
  }
}

resource "tfe_workspace_run" "ws_run_child" {
  organization = tfe_organization.test-organization
  workspace    = tfe_workspace.child.name
  depends_on   = [tfe_workspace_run.ws_run_parent]

  apply {
    retry_attempts = 5
    retry_backoff_min = 5
  }

  destroy {
    retry_attempts = 3
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
  organization = tfe_organization.test-organization
  workspace    = tfe_workspace.parent.name

  apply {
    manual_confirm = true
  }

  destroy {
    manual_confirm = true
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
  organization = tfe_organization.test-organization
  workspace    = tfe_workspace.parent.name

  apply {
    retry = false
  }

  destroy {
    retry = false
  }
}

```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) Name of the workspace to execute the run.
* `organization` - (Optional) Name of the Terraform Cloud organization. If omitted, organization must be defined in the provider config.
* `apply` - (Optional) Settings for the workspace's apply run during creation.
* `destroy` - (Optional) Settings for the workspace's destroy run during destruction.

Both `apply` and `destroy` block supports:

* `manual_confirm` - (Optional) If set to true a human will have to manually confirm a plan to start an apply. If set to false, this resource will auto confirm the plan. The exception is the case of policy check soft-failed where a human has to perform an override by manually confirming the plan even though `manual_confirm` is set to false. Defaults to `false`.
* `retry` - (Optional) Whether or not to retry on plan or apply errors. When set to true, `retry_attempts` must also be greater than zero inorder for retries to happen. Defaults to `true`.
* `retry_attempts` - (Optional) The number to retry attempts made after an initial error. Defaults to `3`.
* `retry_backoff_min` - (Optional) The minimum time in seconds to backoff before attempting a retry. Defaults to `1`.
* `retry_backoff_max` - (Optional) The maximum time in seconds to backoff before attempting a retry. Defaults to `30`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the run created by this resource.

