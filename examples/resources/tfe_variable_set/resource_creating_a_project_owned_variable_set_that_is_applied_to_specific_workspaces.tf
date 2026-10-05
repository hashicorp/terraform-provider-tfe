# Creating a project-owned variable set that is applied to specific workspaces

resource "tfe_project" "test" {
  organization = tfe_organization.example.name
  name         = "workspace-project"
}

resource "tfe_workspace" "test" {
  name         = "project-workspace"
  organization = tfe_organization.example.name
  project_id   = tfe_project.test.id
}

resource "tfe_variable_set" "test" {
  name              = "Project-owned Varset"
  description       = "Varset that is owned and managed by a project."
  organization      = tfe_organization.example.name
  parent_project_id = tfe_project.test.id
}

resource "tfe_workspace_variable_set" "test" {
  workspace_id    = tfe_workspace.test.id
  variable_set_id = tfe_variable_set.test.id
}
