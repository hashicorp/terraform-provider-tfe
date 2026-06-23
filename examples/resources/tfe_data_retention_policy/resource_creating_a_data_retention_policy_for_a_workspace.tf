resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test-workspace" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_data_retention_policy" "foobar" {
  workspace_id = tfe_workspace.test-workspace.id

  delete_older_than {
    days = 42
  }
}
