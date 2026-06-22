resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_policy_set" "test" {
  name         = "my-policy-set"
  description  = "Some description."
  organization = tfe_organization.test.name
  global       = true
}

resource "tfe_tag_policy_set_exclusion" "test" {
  policy_set_id = tfe_policy_set.test.id
  key           = "env"
  value         = "staging"
}
