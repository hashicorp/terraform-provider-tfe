resource "tfe_policy_set" "test" {
  name          = "my-policy-set"
  description   = "A brand new policy set"
  organization  = "my-org-name"
  policies_path = "policies/my-policy-set"
  workspace_ids = [tfe_workspace.test.id]

  vcs_repo {
    identifier         = "my-org-name/my-policy-set-repository"
    branch             = "main"
    ingress_submodules = false
    oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  }
}