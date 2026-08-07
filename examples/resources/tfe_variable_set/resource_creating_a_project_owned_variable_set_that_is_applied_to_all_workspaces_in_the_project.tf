# Creating a project-owned variable set that is applied to all workspaces in the project

resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_project" "test" {
  organization = tfe_organization.test.name
  name         = "projectname"
}

resource "tfe_variable_set" "test" {
  name              = "Project-owned Varset"
  description       = "Varset that is owned and managed by a project."
  organization      = tfe_organization.test.name
  parent_project_id = tfe_project.test.id
}

resource "tfe_project_variable_set" "test" {
  project_id      = tfe_project.test.id
  variable_set_id = tfe_variable_set.test.id
}
