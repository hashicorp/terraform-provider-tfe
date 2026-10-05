# Basic usage

resource "tfe_policy_set" "test" {
  name         = "my-policy-set"
  description  = "Some description."
  organization = tfe_organization.example.name
  global       = true
}

resource "tfe_project_policy_set_exclusion" "test" {
  policy_set_id = tfe_policy_set.test.id
  project_id    = tfe_project.example.id
}
