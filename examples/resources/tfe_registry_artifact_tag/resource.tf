# Basic usage

resource "tfe_registry_module" "example" {
  organization    = tfe_organization.example.name
  name            = "example_module"
  module_provider = "aws"
  registry_name   = "private"
}

resource "tfe_registry_artifact_tag" "example" {
  artifact = {
    type = "registry-module"
    id   = tfe_registry_module.example.id
  }

  tags = [
    {
      key   = "environment"
      value = "production"
    },
    {
      key   = "team"
      value = "platform"
    },
  ]
}
