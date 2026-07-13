# Create private registry module without VCS

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_module" "test-private-registry-module" {
  organization    = tfe_organization.test-organization.name
  module_provider = "my_provider"
  name            = "another_test_module"
  registry_name   = "private"
}
