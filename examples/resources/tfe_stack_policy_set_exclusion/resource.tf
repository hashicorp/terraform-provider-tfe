# Basic usage

resource "tfe_stack" "test" {
  name       = "my-stack-name"
  project_id = tfe_project.example.id
}

resource "tfe_policy_set" "test" {
  name         = "my-policy-set"
  description  = "Some description."
  organization = tfe_organization.example.name
  kind         = "tfpolicy"
  global       = true
}

resource "tfe_stack_policy_set_exclusion" "test" {
  policy_set_id = tfe_policy_set.test.id
  stack_id      = tfe_stack.test.id
}
