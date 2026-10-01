# Basic usage

data "tfe_registry_module" "example" {
  organization    = "my-org-name"
  name            = "my-module"
  module_provider = "aws"
}

data "tfe_registry_artifact_tags" "example" {
  artifact = {
    type = "registry-module"
    id   = data.tfe_registry_module.example.id
  }
}

output "module_tags" {
  value = data.tfe_registry_artifact_tags.example.tags
}
