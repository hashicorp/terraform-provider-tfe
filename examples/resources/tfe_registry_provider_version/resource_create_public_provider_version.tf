# Create a public provider version

data "tfe_organization" "example" {
  name = "my-org-name"
}

data "tfe_registry_gpg_key" "example" {
  organization = data.tfe_organization.example.name
  id           = "ABCDEF1234567890"
}

data "tfe_registry_provider" "example" {
  organization  = data.tfe_organization.example.name
  registry_name = "public"
  namespace     = "hashicorp"
  name          = "aws"
}

resource "tfe_registry_provider_version" "example" {
  organization  = data.tfe_organization.example.name
  registry_name = "public"
  namespace     = "hashicorp"
  name          = data.tfe_registry_provider.example.name
  version       = "5.0.0"
  key_id        = data.tfe_registry_gpg_key.example.id
  protocols     = ["5.0", "6.0"]
}
