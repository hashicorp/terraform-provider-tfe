# A private provider

resource "tfe_registry_provider" "example" {
  organization = "my-org-name"
  name         = "my-provider"
}
