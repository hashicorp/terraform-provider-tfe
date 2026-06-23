resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "time_rotating" "example" {
  rotation_days = 30
}

resource "tfe_team_token" "test" {
  team_id     = tfe_team.test.id
  description = "my team token"
  expired_at  = time_rotating.example.rotation_rfc3339
}
