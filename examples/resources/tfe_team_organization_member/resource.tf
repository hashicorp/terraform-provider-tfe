resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_organization_membership" "test" {
  organization = "my-org-name"
  email = "example@hashicorp.com"
}

resource "tfe_team_organization_member" "test" {
  team_id = tfe_team.test.id
  organization_membership_id = tfe_organization_membership.test.id
}