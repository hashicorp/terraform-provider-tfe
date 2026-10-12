# Basic usage

resource "tfe_policy_set" "test" {
  name         = "my-policy-set"
  description  = "Some description."
  organization = tfe_organization.example.name
  kind         = "tfpolicy"
}

resource "tfe_stack_policy_set" "test" {
  policy_set_id = tfe_policy_set.test.id
  stack_id      = tfe_stack.example.id
}
