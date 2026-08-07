# Basic usage

data "tfe_team_project_access" "test" {
  team_id    = "my-team-id"
  project_id = "my-project-id"
}
