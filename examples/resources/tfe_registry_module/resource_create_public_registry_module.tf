# Create public registry module

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_module" "test-public-registry-module" {
  organization    = tfe_organization.test-organization.name
  namespace       = "terraform-aws-modules"
  module_provider = "aws"
  name            = "vpc"
  registry_name   = "public"
}
