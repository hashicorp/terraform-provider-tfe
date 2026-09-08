# Basic usage

resource "tfe_ip_allowlist" "example" {
  organization      = "my-org-name"
  name              = "agent-pool-network"
  description       = "Allowlist enforced for agent pools"
  enforcement_scope = "all_agent_pools"

  cidr_range = [
    {
      range       = "10.0.0.0/16"
      description = "Corporate LAN"
    },
    {
      range       = "192.168.1.0/24"
      description = "VPN"
      enabled     = false
    },
  ]
}
