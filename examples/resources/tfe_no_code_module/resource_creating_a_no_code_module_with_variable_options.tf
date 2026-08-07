# Creating a no-code module with variable options

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
  version_pin     = "~> 1.1"

  variable_options {
    name    = "ami"
    type    = "string"
    options = ["ami-0", "ami-1", "ami-2"]
  }

  variable_options {
    name    = "region"
    type    = "string"
    options = ["us-east-1", "us-east-2", "us-west-1"]
  }
}
