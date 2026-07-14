# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_organization" "test_org" {
  name            = "my-organization"
  email           = "admin@example.com"
  max_ttl_enabled = true # Enable the max TTL policy feature
}

resource "tfe_org_max_token_ttl_policy" "token_ttl_policy" {
  organization              = tfe_organization.test_org.name
  org_token_max_ttl         = "0.5h"
  user_token_max_ttl        = "2.5y"
  team_token_max_ttl        = "3w"
  audit_trail_token_max_ttl = "6mo"
}
