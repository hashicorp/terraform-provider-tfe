# With tags

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_project" "test" {
  organization = tfe_organization.test-organization.name
  name         = "projectname"
  tags = {
    cost_center = "infrastructure"
    team        = "platform"
  }
}
