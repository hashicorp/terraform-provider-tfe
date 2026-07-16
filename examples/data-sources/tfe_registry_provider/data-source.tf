# A private provider

data "tfe_registry_provider" "example" {
  organization = "my-org-name"
  name         = "my-provider"
}
