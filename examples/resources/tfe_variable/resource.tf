# Basic usage for workspaces

resource "tfe_variable" "test" {
  key          = "my_key_name"
  value        = "my_value_name"
  category     = "terraform"
  workspace_id = tfe_workspace.example.id
  description  = "a useful description"
}
