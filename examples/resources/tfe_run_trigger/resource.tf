resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test-workspace" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.id
}

resource "tfe_workspace" "test-sourceable" {
  name         = "my-sourceable-workspace-name"
  organization = tfe_organization.test-organization.id
}

resource "tfe_run_trigger" "test" {
  workspace_id  = tfe_workspace.test-workspace.id
  sourceable_id = tfe_workspace.test-sourceable.id
}