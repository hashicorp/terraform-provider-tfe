resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_team_member" "test" {
  team_id  = tfe_team.test.id
  username = "sander"
}