# Wait for the Latest Configuration Version

resource "tfe_workspace" "ws" {
  name         = "example-workspace"
  organization = "my-organization"
}

resource "tfe_variable" "var" {
  key          = "my_key"
  value        = "my_value"
  category     = "terraform"
  workspace_id = tfe_workspace.ws.id

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.tfe_query_run.wait_for_latest]
    }
  }
}

action "tfe_query_run" "wait_for_latest" {
  config {
    workspace_id                  = tfe_workspace.ws.id
    wait_for_latest_configuration = true

    variables = {
      "animals" = "5"
    }
  }
}
