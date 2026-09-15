# Create public registry module

resource "tfe_registry_module" "test-public-registry-module" {
  organization    = tfe_organization.example.name
  namespace       = "terraform-aws-modules"
  module_provider = "aws"
  name            = "vpc"
  registry_name   = "public"
}
