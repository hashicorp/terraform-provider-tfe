# Retrieve a specific platform for a provider version

data "tfe_registry_provider_version_platform" "example" {
  organization = "my-org-name"
  name         = "my-provider"
  version      = "1.0.0"
  os_arch      = "linux_amd64"
}
