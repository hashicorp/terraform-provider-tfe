# Basic usage

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@example.com"
}

resource "tfe_project" "test" {
  organization = tfe_organization.test-organization.name
  name         = "projectname"
}
