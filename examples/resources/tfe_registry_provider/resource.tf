# Create private provider

resource "tfe_registry_provider" "example" {
  organization = tfe_organization.example.name

  name = "my-provider"
}
