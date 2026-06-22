resource "tfe_policy_set" "test" {
  name                = "my-policy-set"
  description         = "A brand new policy set"
  organization        = "my-org-name"
  kind                = "sentinel"
  agent_enabled       = "true"
  policy_tool_version = "0.24.1"
  policy_ids          = [tfe_sentinel_policy.test.id]
  workspace_ids       = [tfe_workspace.test.id]
}
