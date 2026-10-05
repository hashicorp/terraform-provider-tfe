# Basic usage

resource "tfe_policy_set" "example" {
  name         = "my-policy-set"
  description  = "Example fixture policy set."
  organization = tfe_organization.example.name
}

resource "tfe_project_policy_set" "test" {
  policy_set_id = tfe_policy_set.example.id
  project_id    = tfe_project.example.id
}
