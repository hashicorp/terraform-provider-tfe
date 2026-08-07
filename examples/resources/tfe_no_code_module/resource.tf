# Basic usage

resource "tfe_organization" "foobar" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_module" "foobar" {
  organization    = tfe_organization.foobar.id
  module_provider = "my_provider"
  name            = "test_module"
}

resource "tfe_no_code_module" "foobar" {
  organization    = tfe_organization.foobar.id
  registry_module = tfe_registry_module.foobar.id
}
