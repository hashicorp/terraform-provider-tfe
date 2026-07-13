# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_org_max_token_ttl_policy" "example" {
  organization = "my-org-name"
}

output "org_token_ttl_ms" {
  value = data.tfe_org_max_token_ttl_policy.example.org_token_max_ttl_ms
}
