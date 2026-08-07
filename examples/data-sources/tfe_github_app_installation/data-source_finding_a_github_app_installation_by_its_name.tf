# Finding a Github App Installation by its name

data "tfe_github_app_installation" "gha_installation" {
  name = "github_username_or_organization"
}
