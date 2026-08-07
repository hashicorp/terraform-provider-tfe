# Basic usage

resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_stack" "test" {
  project_id = tfe_organization.test.default_project_id
  name       = "my-stack-name"
}

resource "tfe_variable_set" "test" {
  name         = "Test Varset"
  description  = "Some description."
  organization = tfe_organization.test.id
}

resource "tfe_stack_variable_set" "test" {
  stack_id        = tfe_stack.test.id
  variable_set_id = tfe_variable_set.test.id
}
