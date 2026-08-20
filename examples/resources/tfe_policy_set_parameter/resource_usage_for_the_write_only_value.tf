# Usage for the write-only value

resource "tfe_policy_set" "example" {
  name         = "my-policy-set"
  description  = "Example fixture policy set."
  organization = tfe_organization.example.name
}

variable "session_token" {
  type      = string
  ephemeral = true
}

resource "tfe_policy_set_parameter" "test" {
  key              = "my_key_name"
  value_wo         = var.session_token
  value_wo_version = 1
  policy_set_id    = tfe_policy_set.example.id
}
