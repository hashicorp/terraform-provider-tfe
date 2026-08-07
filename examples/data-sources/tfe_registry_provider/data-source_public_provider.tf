# A public provider

data "tfe_registry_provider" "example" {
  organization  = "my-org-name"
  registry_name = "public"
  namespace     = "hashicorp"
  name          = "aws"
}
