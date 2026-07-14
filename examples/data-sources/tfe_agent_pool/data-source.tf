# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_agent_pool" "test" {
  name         = "my-agent-pool-name"
  organization = "my-org-name"
}
