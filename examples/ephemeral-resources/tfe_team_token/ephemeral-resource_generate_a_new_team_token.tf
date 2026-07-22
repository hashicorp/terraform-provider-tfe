# Generate a new team token

resource "tfe_team" "example" {
  organization = "my-org-name"
  name         = "my-team-name"
}

ephemeral "tfe_team_token" "example" {
  team_id = tfe_team.example.id
}
