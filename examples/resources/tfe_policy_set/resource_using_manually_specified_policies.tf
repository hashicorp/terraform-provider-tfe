# Using manually-specified policies

resource "tfe_policy" "test" {
  name         = "passing-policy"
  description  = "This policy always passes"
  organization = "my-org-name"
  kind         = "sentinel"
  policy       = "main = rule { true }"
  enforce_mode = "hard-mandatory"
}

resource "tfe_policy_set" "test" {
  name                = "manual-policy-set"
  description         = "A brand new policy set"
  organization        = "my-org-name"
  kind                = "sentinel"
  agent_enabled       = "true"
  policy_tool_version = "0.24.1"
  policy_ids          = [tfe_policy.test.id]
  workspace_ids       = [tfe_workspace.example.id]
}
