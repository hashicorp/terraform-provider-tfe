data "tfe_workspace_ids" "app-frontend" {
  names        = ["app-frontend-prod", "app-frontend-dev1", "app-frontend-staging"]
  organization = "my-org-name"
}

data "tfe_workspace_ids" "all" {
  names        = ["*"]
  organization = "my-org-name"
}

data "tfe_workspace_ids" "prod-apps" {
  tag_names    = ["prod", "app", "aws"]
  organization = "my-org-name"
}

data "tfe_workspace_ids" "prod-only" {
  tag_names    = ["prod"]
  exclude_tags = ["app"]
  organization = "my-org-name"
}