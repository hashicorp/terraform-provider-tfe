resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_module" "test-no-code-provisioning-registry-module" {
  organization    = tfe_organization.test-organization.name
  namespace       = "terraform-aws-modules"
  module_provider = "aws"
  name            = "vpc"
  registry_name   = "public"
}

resource "tfe_no_code_module" "foobar" {
  organization    = tfe_organization.test-organization.id
  registry_module = tfe_registry_module.test-no-code-provisioning-registry-module.id
}
