data "tfe_slug" "test" {
  // point to the local directory where the policies are located.
  source_path = "policies/my-policy-set"
}

resource "tfe_policy_set" "test" {
  name          = "my-policy-set"
  description   = "A brand new policy set"
  organization  = "my-org-name"
  workspace_ids = [tfe_workspace.test.id]

  // reference the tfe_slug data source.
  slug = data.tfe_slug.test
}