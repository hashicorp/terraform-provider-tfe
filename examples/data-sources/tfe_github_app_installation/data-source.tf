# Finding a Github App Installation by its installation ID

data "tfe_github_app_installation" "gha_installation" {
  installation_id = 12345678
}
