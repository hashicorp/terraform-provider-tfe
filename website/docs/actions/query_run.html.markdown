---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_query_run"
description: |-
  Creates a query run in a HCP Terraform or Terraform Enterprise workspace.
---

# Action: tfe_query_run

Initiates a query run within a specified workspace. This action allows you to execute a query on a workspace either against a specific configuration version or by waiting for the latest configuration version to be available.

## Example Usage

### Using a Specific Configuration Version ID

```terraform
resource "tfe_workspace" "example" {
  name         = "example-workspace"
  organization = "my-organization"
}

resource "tfe_variable" "example" {
  key          = "my_key"
  value        = "my_value"
  category     = "terraform"
  workspace_id = tfe_workspace.example.id

  # Trigger the query run after the variable is created or updated
  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.tfe_query_run.with_cv_id]
    }
  }
}

action "tfe_query_run" "with_cv_id" {
  config {
    workspace_id             = tfe_workspace.example.id
    configuration_version_id = "cv-ntv3HbhJqvFzamy7"

    variables = {
      "animals" = "5"
    }
  }
}
```

### Wait for the Latest Configuration Version

```terraform
resource "tfe_workspace" "example" {
  name         = "example-workspace"
  organization = "my-organization"
}

resource "tfe_variable" "example" {
  key          = "my_key"
  value        = "my_value"
  category     = "terraform"
  workspace_id = tfe_workspace.example.id

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.tfe_query_run.wait_for_latest]
    }
  }
}

action "tfe_query_run" "wait_for_latest" {
  config {
    workspace_id                  = tfe_workspace.example.id
    wait_for_latest_configuration = true

    variables = {
      "animals" = "5"
    }
  }
}
```

### Invoking the action directly

```sh
terraform apply -invoke=action.tfe_query_run.wait_for_latest
```

## Argument Reference

This action supports the following arguments within the config block:

* `workspace_id` - (Required) The ID of the workspace where the query run will be executed.
* `configuration_version_id` - (Optional) The specific Configuration Version ID to use for the query run (e.g., "cv-ntv3HbhJqvFzamy7"). Exactly one of configuration_version_id or wait_for_latest_configuration must be provided.
* `wait_for_latest_configuration` - (Optional) A boolean flag that, when set to true, tells the action to wait for and use the latest configuration version available in the workspace. Exactly one of wait_for_latest_configuration or configuration_version_id must be provided.
* `variables` - (Optional) A map of key-value string pairs representing variables to pass directly into the query run.

