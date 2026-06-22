resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

// create data retention policy the organization
resource "tfe_data_retention_policy" "foobar" {
  organization = tfe_organization.test-organization.name

  delete_older_than {
    days = 1138
  }
}

resource "tfe_workspace" "test-workspace" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.name
}

// create a policy that prevents automatic deletion of data in the test-workspace
resource "tfe_data_retention_policy" "foobar" {
  workspace_id = tfe_workspace.test-workspace.id

  dont_delete {}
}
