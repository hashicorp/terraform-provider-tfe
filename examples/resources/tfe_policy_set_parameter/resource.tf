# Basic usage

resource "tfe_policy_set" "example" {
  name         = "my-policy-set"
  description  = "Example fixture policy set."
  organization = tfe_organization.example.name
}

resource "tfe_policy_set_parameter" "test" {
  key           = "my_key_name"
  value         = "my_value_name"
  policy_set_id = tfe_policy_set.example.id
}
