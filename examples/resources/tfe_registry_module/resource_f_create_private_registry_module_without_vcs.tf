# Create private registry module without VCS

resource "tfe_registry_module" "test-private-registry-module" {
  organization    = tfe_organization.example.name
  module_provider = "my_provider"
  name            = "another_test_module"
  registry_name   = "private"
}
