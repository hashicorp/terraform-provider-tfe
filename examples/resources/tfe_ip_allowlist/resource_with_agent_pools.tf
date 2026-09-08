# Restrict an allowlist to a specific, existing agent pool.
#
# When enforcement_scope is "selected_agent_pools", the allowlist only applies
# to the agent pools listed in agent_pool_ids. Here we look up an agent pool
# that already exists by name and assign the allowlist to it.

data "tfe_agent_pool" "example" {
  name         = "my-agent-pool-name"
  organization = "my-org-name"
}

resource "tfe_ip_allowlist" "example" {
  organization      = "my-org-name"
  name              = "agent-pool-allowlist"
  description       = "Only enforced for the selected agent pool"
  enforcement_scope = "selected_agent_pools"
  agent_pool_ids    = [data.tfe_agent_pool.example.id]

  cidr_range = [
    {
      range       = "203.0.113.0/24"
      description = "Datacenter egress"
    },
  ]
}
