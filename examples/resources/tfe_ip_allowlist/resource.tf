# Basic usage

resource "tfe_ip_allowlist" "example" {
  organization      = "my-org-name"
  name              = "corporate-network"
  description       = "Allowlist for the corporate network"
  enforcement_scope = "organization"

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
