# Basic usage

resource "tfe_project" "test" {
  organization = tfe_organization.example.name
  name         = "projectname"
}
