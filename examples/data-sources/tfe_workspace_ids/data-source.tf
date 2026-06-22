data "tfe_workspace_ids" "app-frontend" {
  names        = ["app-frontend-prod", "app-frontend-dev1", "app-frontend-staging"]
  organization = "my-org-name"
}

data "tfe_workspace_ids" "all" {
  names        = ["*"]
  organization = "my-org-name"
}

data "tfe_workspace_ids" "dev_env_tags_only" {
  organization = "my-org-name"
  tag_filters {
    include = {
      environment = "dev"
    }
  }
}

data "tfe_workspace_ids" "include_and_exclude" {
  organization = "my-org-name"
  tag_filters {
    include = {
      region = "us-east-1"
    }

    exclude = {
      team = "prodsec"
    }
  }
}

data "tfe_workspace_ids" "exclude_all_matching_key" {
  organization = "my-org-name"
  tag_filters {
    exclude = {
      bad_key = "*"
    }
  }
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
