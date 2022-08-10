data "tfe_variable_set" "test" {
  name         = "my-variable-set-name"
  organization = "my-org-name"
}

data "tfe_variables" "test" {
  variable_set_id = data.tfe_variable_set.test.id
}