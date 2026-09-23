# Basic usage

resource "tfe_ip_allowlist" "example" {
  organization      = "my-org-name"
  name              = "corporate-network"
  enforcement_scope = "organization"

  cidr_range = [
    {
      range       = "10.0.0.0/16"
      description = "Corporate LAN"
    },
  ]
}

data "tfe_ip_allowlist" "example" {
  name         = tfe_ip_allowlist.example.name
  organization = "my-org-name"
}
