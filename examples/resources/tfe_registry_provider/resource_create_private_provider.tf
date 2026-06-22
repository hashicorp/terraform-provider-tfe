resource "tfe_organization" "example" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_provider" "example" {
  organization = tfe_organization.example.name

  name = "my-provider"
}
