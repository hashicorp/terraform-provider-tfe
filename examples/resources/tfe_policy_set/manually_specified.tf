resource "tfe_policy_set" "test" {
  name          = "my-policy-set"
  description   = "A brand new policy set"
  organization  = "my-org-name"
  policy_ids    = [tfe_sentinel_policy.test.id]
  workspace_ids = [tfe_workspace.test.id]
}