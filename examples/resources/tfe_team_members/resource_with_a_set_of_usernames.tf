locals {
  all_usernames = toset([
    "user1",
    "user2",
  ])
}

resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_team_members" "test" {
  team_id   = tfe_team.test.id
  usernames = [for user in local.all_usernames : user]
}
