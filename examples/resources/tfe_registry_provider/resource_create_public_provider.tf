# Create public provider

resource "tfe_registry_provider" "example" {
  organization = tfe_organization.example.name

  registry_name = "public"
  namespace     = "hashicorp"
  name          = "aws"
}
