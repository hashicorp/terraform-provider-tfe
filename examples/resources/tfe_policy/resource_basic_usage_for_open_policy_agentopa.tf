resource "tfe_policy" "test" {
  name         = "my-policy-name"
  description  = "This policy always passes"
  organization = "my-org-name"
  kind         = "opa"
  policy       = "package example rule[\"not allowed\"] { false }"
  query        = "data.example.rule"
  enforce_mode = "mandatory"
}
