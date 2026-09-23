# List all versions of a provider

data "tfe_registry_provider_versions" "example" {
  organization = "my-org-name"
  name         = "my-provider"
}
