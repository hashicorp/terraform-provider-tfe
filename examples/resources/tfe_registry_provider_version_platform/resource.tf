# Create a platform binary for a provider version

data "tfe_organization" "example" {
  name = "my-org-name"
}

data "tfe_registry_provider" "example" {
  organization = data.tfe_organization.example.name
  name         = "my-provider"
}

data "tfe_registry_provider_version" "example" {
  organization = data.tfe_organization.example.name
  name         = data.tfe_registry_provider.example.name
  version      = "1.0.0"
}

resource "tfe_registry_provider_version_platform" "example" {
  organization = data.tfe_organization.example.name
  name         = data.tfe_registry_provider_version.example.name
  version      = data.tfe_registry_provider_version.example.version
  os_arch      = "linux_amd64"
  filename     = "https://releases.example.com/my-provider/1.0.0/my-provider_1.0.0_linux_amd64.zip"
}
