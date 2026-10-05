# Retrieve a specific provider version

data "tfe_registry_provider_version" "example" {
  organization = "my-org-name"
  name         = "my-provider"
  version      = "1.0.0"
}
