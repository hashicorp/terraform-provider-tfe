resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

data "tfe_workspace_ids" "prod-apps" {
  tag_names    = ["prod", "app", "aws"]
  organization = tfe_organization.test.name
}

resource "tfe_variable_set" "test" {
  name          = "Tag Based Varset"
  description   = "Variable set applied to workspaces based on tag."
  organization  = tfe_organization.test.name
}

resource "tfe_workspace_variable_set" "test" {
  for_each        = toset(values(data.tfe_workspace_ids.prod-apps.ids))
  workspace_id    = each.key
  variable_set_id = tfe_variable_set.test.id
}