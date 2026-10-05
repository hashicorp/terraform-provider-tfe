# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage for Open Policy Agent (OPA)

resource "tfe_policy" "test" {
  name         = "my-policy-name"
  description  = "This policy always passes"
  organization = "my-org-name"
  kind         = "opa"
  policy       = "package example rule[\"not allowed\"] { false }"
  query        = "data.example.rule"
  enforce_mode = "mandatory"
}
