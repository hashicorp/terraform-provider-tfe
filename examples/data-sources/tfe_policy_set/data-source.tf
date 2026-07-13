# For workspace policies

data "tfe_policy_set" "test" {
  name         = "my-policy-set-name"
  organization = "my-org-name"
}
