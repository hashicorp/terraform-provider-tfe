resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
  organization_access {
    manage_vcs_settings = true
  }
}