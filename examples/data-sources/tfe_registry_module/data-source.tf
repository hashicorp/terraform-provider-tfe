# Basic usage

data "tfe_registry_module" "example" {
  organization    = "my-organization"
  name            = "no-code-ssm"
  module_provider = "aws"
}
