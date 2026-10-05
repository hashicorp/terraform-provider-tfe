# Create a platform binary for a provider version
resource "tfe_registry_provider" "example" {
  organization = tfe_organization.example.name
  name = "my-provider"
}

resource "tfe_registry_gpg_key" "example" {
  organization = tfe_organization.example.name
  ascii_armor  = file("my-public-key.asc")
}

resource "tfe_registry_provider_version" "example" {
  organization = tfe_registry_provider.example.organization
  name         = tfe_registry_provider.example.name
  version      = "1.0.0"
  key_id       = tfe_registry_gpg_key.example.id
  protocols    = ["5.0"]
}

resource "tfe_registry_provider_version_platform" "example" {
  organization = tfe_registry_provider_version.example.organization
  name         = tfe_registry_provider_version.example.name
  version      = tfe_registry_provider_version.example.version
  os_arch      = "linux_amd64"
  filename     = "https://releases.example.com/my-provider/1.0.0/my-provider_1.0.0_linux_amd64.zip"
}
