# List all platforms for a provider version

data "tfe_registry_provider_version_platforms" "example" {
  organization = "my-org-name"
  name         = "my-provider"
  version      = "1.0.0"
}
