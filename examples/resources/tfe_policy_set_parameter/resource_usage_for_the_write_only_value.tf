variable "session_token" {
  type      = string
  ephemeral = true
}

resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_policy_set" "test" {
  name         = "my-policy-set-name"
  organization = tfe_organization.test.id
}

resource "tfe_policy_set_parameter" "test" {
  key              = "my_key_name"
  value_wo         = var.session_token
  value_wo_version = 1
  policy_set_id    = tfe_policy_set.test.id
}
