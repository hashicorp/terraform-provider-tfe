resource "tfe_sentinel_policy" "test" {
  name         = "my-policy-name"
  description  = "This policy always passes"
  organization = "my-org-name"
  policy       = "main = rule { true }"
  enforce_mode = "hard-mandatory"
}