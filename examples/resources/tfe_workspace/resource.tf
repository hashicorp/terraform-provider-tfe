# Basic usage

resource "tfe_workspace" "test" {
  name         = "user-workspace"
  organization = tfe_organization.example.name
  tags = {
    environment = "prod"
    team_owner  = "my-team"
  }
}
