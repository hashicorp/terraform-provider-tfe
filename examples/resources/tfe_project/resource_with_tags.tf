# With tags

resource "tfe_project" "test" {
  organization = tfe_organization.example.name
  name         = "projectname"
  tags = {
    cost_center = "infrastructure"
    team        = "platform"
  }
}
