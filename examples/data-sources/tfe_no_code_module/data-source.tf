resource "tfe_no_code_module" "foobar" {
  organization    = tfe_organization.foobar.id
  registry_module = tfe_registry_module.foobar.id
}

data "tfe_no_code_module" "foobar" {
  id = tfe_no_code_module.foobar.id
}
