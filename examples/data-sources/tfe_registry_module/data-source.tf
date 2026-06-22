data "tfe_registry_module" "example" {
  organization    = var.organization_name
  name            = "no-code-ssm"
  module_provider = "aws"
}
