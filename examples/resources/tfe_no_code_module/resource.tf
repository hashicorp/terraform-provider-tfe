# Basic usage

resource "tfe_registry_module" "foobar" {
  organization    = tfe_organization.example.id
  module_provider = "my_provider"
  name            = "test_module"
}

resource "tfe_no_code_module" "foobar" {
  organization    = tfe_organization.example.id
  registry_module = tfe_registry_module.foobar.id
}
