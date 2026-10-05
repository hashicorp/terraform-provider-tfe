# Create a private provider version

resource "tfe_registry_provider_version" "example" {
  organization = tfe_registry_provider.example.organization
  name         = tfe_registry_provider.example.name
  version      = "1.0.0"
  key_id       = tfe_registry_gpg_key.example.id
  protocols    = ["5.0"]
}
