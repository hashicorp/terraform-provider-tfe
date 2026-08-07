# Using manually-specified policies

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_policy" "test" {
  name         = "my-policy-name"
  description  = "This policy always passes"
  organization = "my-org-name"
  kind         = "sentinel"
  policy       = "main = rule { true }"
  enforce_mode = "hard-mandatory"
}

resource "tfe_policy_set" "test" {
  name                = "my-policy-set"
  description         = "A brand new policy set"
  organization        = "my-org-name"
  kind                = "sentinel"
  agent_enabled       = "true"
  policy_tool_version = "0.24.1"
  policy_ids          = [tfe_policy.test.id]
  workspace_ids       = [tfe_workspace.test.id]
}
