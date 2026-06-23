resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_project" "test" {
  organization = tfe_organization.test.name
  name         = "projectname"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.name
  project_id   = tfe_project.test.id
}

resource "tfe_variable_set" "test" {
  name              = "Project-owned Varset"
  description       = "Varset that is owned and managed by a project."
  organization      = tfe_organization.test.name
  parent_project_id = tfe_project.test.id
}

resource "tfe_workspace_variable_set" "test" {
  workspace_id    = tfe_workspace.test.id
  variable_set_id = tfe_variable_set.test.id
}
