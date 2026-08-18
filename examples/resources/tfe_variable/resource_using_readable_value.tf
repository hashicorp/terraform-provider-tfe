# Using readable_value

# While the `value` field may be referenced in other resources, for safety it is always treated as sensitive. This means that it will always be redacted from plan outputs, and any other resource attributes which depend on it will also be redacted.

# The `readable_value` attribute is not sensitive, and will not be redacted; instead, it will be null if the variable is sensitive. This allows other resources to reference it, while keeping their plan outputs readable.

resource "tfe_variable" "sensitive_var" {
  key          = "sensitive_key"
  value        = "sensitive_value" // this will be redacted from plan outputs
  category     = "terraform"
  workspace_id = tfe_workspace.workspace.id
  sensitive    = true
}

resource "tfe_variable" "visible_var" {
  key          = "visible_key"
  value        = "visible_value" // this will be redacted from plan outputs
  category     = "terraform"
  workspace_id = tfe_workspace.workspace.id
  sensitive    = false
}

resource "tfe_workspace" "workspace" {
  name         = "my-workspace"
  organization = "organization name"
}

resource "tfe_workspace" "sensitive_workspace" {
  name         = "workspace-${tfe_variable.sensitive_var.value}" // this will be redacted from plan outputs
  organization = "organization name"
}

resource "tfe_workspace" "visible_workspace" {
  name         = "workspace-${tfe_variable.visible_var.readable_value}" // this will not be redacted from plan outputs
  organization = "organization name"
}

