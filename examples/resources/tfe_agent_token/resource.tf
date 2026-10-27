# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_agent_pool" "test-agent-pool" {
  name         = "my-agent-pool-name"
  organization = tfe_organization.example.id
}

resource "tfe_agent_token" "test-agent-token" {
  agent_pool_id = tfe_agent_pool.test-agent-pool.id
  description   = "my-agent-token-name"
}
