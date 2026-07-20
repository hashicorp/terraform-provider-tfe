# Basic usage for Sentinel

resource "tfe_policy" "test" {
  name         = "my-policy-name"
  description  = "This policy always passes"
  organization = "my-org-name"
  kind         = "sentinel"
  policy       = "main = rule { true }"
  enforce_mode = "hard-mandatory"
}
