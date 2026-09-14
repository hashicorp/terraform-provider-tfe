# Basic usage

data "tfe_registry_module" "example" {
  organization    = "my-org-name"
  name            = "my-module"
  module_provider = "aws"
}

resource "tfe_registry_artifact_tag" "example" {
  artifact = {
    type = "registry-module"
    id   = data.tfe_registry_module.example.id
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
