resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

data "tfe_github_app_installation" "gha_installation" {
  name = "YOUR_GH_NAME"
}

resource "tfe_registry_module" "petstore" {
  organization = tfe_organization.test-organization.name
  vcs_repo {
    display_identifier         = "GH_NAME/REPO_NAME"
    identifier                 = "GH_NAME/REPO_NAME"
    github_app_installation_id = data.tfe_github_app_installation.gha_installation.id
  }
}
