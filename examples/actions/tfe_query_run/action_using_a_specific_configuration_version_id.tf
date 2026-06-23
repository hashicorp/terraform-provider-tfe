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
