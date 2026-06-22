resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_project" "test" {
  name         = "my-project-name"
  organization = tfe_organization.test.name
}

resource "tfe_policy_set" "test" {
  name         = "my-policy-set"
  description  = "Some description."
  organization = tfe_organization.test.name
}

resource "tfe_project_policy_set" "test" {
  policy_set_id = tfe_policy_set.test.id
  project_id    = tfe_project.test.id
}
