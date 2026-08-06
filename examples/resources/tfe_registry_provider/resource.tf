# Create private provider

resource "tfe_organization" "example" {
  name  = "my-org-name"
  email = "admin@example.com"
}

resource "tfe_registry_provider" "example" {
  organization = tfe_organization.example.name

  name = "my-provider"
}
