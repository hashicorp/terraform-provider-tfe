# Creating a data retention policy for an organization

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_data_retention_policy" "foobar" {
  organization = tfe_organization.test-organization.name

  delete_older_than {
    days = 1138
  }
}
