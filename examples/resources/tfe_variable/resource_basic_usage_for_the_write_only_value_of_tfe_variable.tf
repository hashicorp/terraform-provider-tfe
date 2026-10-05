# Basic usage for the write-only value of tfe_variable

variable "session_token" {
  type      = string
  ephemeral = true
}

resource "tfe_variable" "test" {
  key          = "my_key_name"
  value_wo     = var.session_token
  category     = "terraform"
  workspace_id = tfe_workspace.example.id
  description  = "a useful description"
}
