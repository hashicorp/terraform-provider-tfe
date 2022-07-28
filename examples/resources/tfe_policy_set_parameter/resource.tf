resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_policy_set" "test" {
  name         = "my-policy-set-name"
  organization = tfe_organization.test.id
}

resource "tfe_policy_set_parameter" "test" {
  key          = "my_key_name"
  value        = "my_value_name"
  policy_set_id = tfe_policy_set.test.id
}