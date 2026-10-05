# Create a platform binary for a provider version

resource "tfe_registry_provider_version_platform" "example" {
  organization = tfe_registry_provider_version.example.organization
  name         = tfe_registry_provider_version.example.name
  version      = tfe_registry_provider_version.example.version
  os_arch      = "linux_amd64"
  filename     = "https://releases.example.com/my-provider/1.0.0/my-provider_1.0.0_linux_amd64.zip"
}
